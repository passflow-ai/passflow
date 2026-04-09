package acf

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystemToolProvider_Read(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("hello world"), 0644)

	stp := NewSystemToolProvider(dir, nil)
	result, err := stp.Execute(context.Background(), &ToolCall{
		Tool:   "Read",
		Params: map[string]interface{}{"file_path": testFile},
		CallID: "c1",
	})
	require.NoError(t, err)
	assert.Contains(t, result.Output, "hello world")
	assert.Empty(t, result.Error)
}

func TestSystemToolProvider_Read_PathTraversal(t *testing.T) {
	dir := t.TempDir()
	stp := NewSystemToolProvider(dir, nil)
	result, err := stp.Execute(context.Background(), &ToolCall{
		Tool:   "Read",
		Params: map[string]interface{}{"file_path": "/etc/passwd"},
		CallID: "c1",
	})
	require.NoError(t, err)
	assert.Contains(t, result.Error, "outside work directory")
}

func TestSystemToolProvider_Write(t *testing.T) {
	dir := t.TempDir()
	stp := NewSystemToolProvider(dir, nil)
	outFile := filepath.Join(dir, "out.txt")
	result, err := stp.Execute(context.Background(), &ToolCall{
		Tool:   "Write",
		Params: map[string]interface{}{"file_path": outFile, "content": "written"},
		CallID: "c1",
	})
	require.NoError(t, err)
	assert.Empty(t, result.Error)

	data, _ := os.ReadFile(outFile)
	assert.Equal(t, "written", string(data))
}

func TestSystemToolProvider_Write_CreatesParentDirs(t *testing.T) {
	dir := t.TempDir()
	stp := NewSystemToolProvider(dir, nil)
	outFile := filepath.Join(dir, "subdir", "nested", "file.txt")
	result, err := stp.Execute(context.Background(), &ToolCall{
		Tool:   "Write",
		Params: map[string]interface{}{"file_path": outFile, "content": "nested"},
		CallID: "c2",
	})
	require.NoError(t, err)
	assert.Empty(t, result.Error)

	data, _ := os.ReadFile(outFile)
	assert.Equal(t, "nested", string(data))
}

func TestSystemToolProvider_Edit(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "edit.txt")
	os.WriteFile(testFile, []byte("foo bar baz"), 0644)

	stp := NewSystemToolProvider(dir, nil)
	result, err := stp.Execute(context.Background(), &ToolCall{
		Tool: "Edit",
		Params: map[string]interface{}{
			"file_path":  testFile,
			"old_string": "bar",
			"new_string": "QUX",
		},
		CallID: "c1",
	})
	require.NoError(t, err)
	assert.Empty(t, result.Error)
	assert.Contains(t, result.Output, "Edit applied")

	data, _ := os.ReadFile(testFile)
	assert.Equal(t, "foo QUX baz", string(data))
}

func TestSystemToolProvider_Edit_OldStringNotFound(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "edit.txt")
	os.WriteFile(testFile, []byte("hello world"), 0644)

	stp := NewSystemToolProvider(dir, nil)
	result, err := stp.Execute(context.Background(), &ToolCall{
		Tool: "Edit",
		Params: map[string]interface{}{
			"file_path":  testFile,
			"old_string": "NOTHERE",
			"new_string": "x",
		},
		CallID: "c1",
	})
	require.NoError(t, err)
	assert.Contains(t, result.Error, "old_string not found")
}

func TestSystemToolProvider_Grep(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.go"), []byte("func main() {}"), 0644)
	os.WriteFile(filepath.Join(dir, "b.go"), []byte("func helper() {}"), 0644)

	stp := NewSystemToolProvider(dir, nil)
	result, err := stp.Execute(context.Background(), &ToolCall{
		Tool:   "Grep",
		Params: map[string]interface{}{"pattern": "func main"},
		CallID: "c1",
	})
	require.NoError(t, err)
	assert.Contains(t, result.Output, "a.go")
	assert.NotContains(t, result.Output, "b.go")
}

func TestSystemToolProvider_Grep_NoMatches(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.go"), []byte("func main() {}"), 0644)

	stp := NewSystemToolProvider(dir, nil)
	result, err := stp.Execute(context.Background(), &ToolCall{
		Tool:   "Grep",
		Params: map[string]interface{}{"pattern": "NOMATCH"},
		CallID: "c1",
	})
	require.NoError(t, err)
	assert.Contains(t, result.Output, "No matches found")
}

func TestSystemToolProvider_Grep_InvalidRegex(t *testing.T) {
	stp := NewSystemToolProvider(t.TempDir(), nil)
	result, err := stp.Execute(context.Background(), &ToolCall{
		Tool:   "Grep",
		Params: map[string]interface{}{"pattern": "[invalid"},
		CallID: "c1",
	})
	require.NoError(t, err)
	assert.Contains(t, result.Error, "invalid regex")
}

func TestSystemToolProvider_Glob(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.go"), []byte(""), 0644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte(""), 0644)

	stp := NewSystemToolProvider(dir, nil)
	result, err := stp.Execute(context.Background(), &ToolCall{
		Tool:   "Glob",
		Params: map[string]interface{}{"pattern": "*.go"},
		CallID: "c1",
	})
	require.NoError(t, err)
	assert.Contains(t, result.Output, "a.go")
	assert.NotContains(t, result.Output, "b.txt")
}

func TestSystemToolProvider_Glob_NoMatches(t *testing.T) {
	dir := t.TempDir()
	stp := NewSystemToolProvider(dir, nil)
	result, err := stp.Execute(context.Background(), &ToolCall{
		Tool:   "Glob",
		Params: map[string]interface{}{"pattern": "*.rs"},
		CallID: "c1",
	})
	require.NoError(t, err)
	assert.Contains(t, result.Output, "No matches found")
}

func TestSystemToolProvider_LS(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte(""), 0644)
	os.Mkdir(filepath.Join(dir, "subdir"), 0755)

	stp := NewSystemToolProvider(dir, nil)
	result, err := stp.Execute(context.Background(), &ToolCall{
		Tool:   "LS",
		Params: map[string]interface{}{},
		CallID: "c1",
	})
	require.NoError(t, err)
	assert.Contains(t, result.Output, "file.txt")
	assert.Contains(t, result.Output, "subdir")
}

func TestSystemToolProvider_LS_PathTraversal(t *testing.T) {
	dir := t.TempDir()
	stp := NewSystemToolProvider(dir, nil)
	result, err := stp.Execute(context.Background(), &ToolCall{
		Tool:   "LS",
		Params: map[string]interface{}{"path": "/tmp"},
		CallID: "c1",
	})
	require.NoError(t, err)
	assert.Contains(t, result.Error, "outside work directory")
}

func TestSystemToolProvider_UnknownTool(t *testing.T) {
	stp := NewSystemToolProvider(t.TempDir(), nil)
	result, err := stp.Execute(context.Background(), &ToolCall{
		Tool: "Unknown", CallID: "c1",
	})
	require.NoError(t, err)
	assert.Contains(t, result.Error, "unknown system tool")
}

func TestSystemToolProvider_PolicyDenied(t *testing.T) {
	policy := &Policy{DeniedTools: []string{"Bash"}}
	stp := NewSystemToolProvider(t.TempDir(), policy)
	result, err := stp.Execute(context.Background(), &ToolCall{
		Tool: "Bash", Params: map[string]interface{}{"command": "ls"}, CallID: "c1",
	})
	require.NoError(t, err)
	assert.Contains(t, result.Error, "denied by workspace policy")
}

func TestSystemToolProvider_PolicyAllowedList(t *testing.T) {
	policy := &Policy{AllowedTools: []string{"Read", "Grep"}}
	stp := NewSystemToolProvider(t.TempDir(), policy)

	// Write is not in the allowed list — should be denied.
	result, err := stp.Execute(context.Background(), &ToolCall{
		Tool:   "Write",
		Params: map[string]interface{}{"file_path": "x", "content": "y"},
		CallID: "c1",
	})
	require.NoError(t, err)
	assert.Contains(t, result.Error, "not in allowed list")
}

func TestSystemToolProvider_ResultCallIDPropagated(t *testing.T) {
	stp := NewSystemToolProvider(t.TempDir(), nil)
	result, err := stp.Execute(context.Background(), &ToolCall{
		Tool: "Unknown", CallID: "my-call-id",
	})
	require.NoError(t, err)
	assert.Equal(t, "my-call-id", result.CallID)
}

func TestSystemToolProvider_Bash(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping bash test in short mode")
	}
	dir := t.TempDir()
	stp := NewSystemToolProvider(dir, nil)
	result, err := stp.Execute(context.Background(), &ToolCall{
		Tool:   "Bash",
		Params: map[string]interface{}{"command": "echo hello_from_bash"},
		CallID: "c1",
	})
	require.NoError(t, err)
	assert.Contains(t, result.Output, "hello_from_bash")
	assert.Empty(t, result.Error)
}

func TestSystemToolProvider_Bash_MissingCommand(t *testing.T) {
	stp := NewSystemToolProvider(t.TempDir(), nil)
	result, err := stp.Execute(context.Background(), &ToolCall{
		Tool:   "Bash",
		Params: map[string]interface{}{},
		CallID: "c1",
	})
	require.NoError(t, err)
	assert.Contains(t, result.Error, "command parameter required")
}
