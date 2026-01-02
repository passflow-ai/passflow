const Problem = () => {
  const problems = [
    "Synthetic identity fraud",
    "High manual review costs",
    "Legit users dropping off",
    "Slow time-to-revenue",
  ];

  return (
    <section className="section-padding bg-white">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="max-w-3xl mx-auto text-center">
          <h2 className="text-3xl md:text-4xl lg:text-5xl font-bold text-[#0f172a] mb-6">
            Fraud and fake users kill growth.
          </h2>
          <p className="text-lg text-[#475569] mb-10">
            Every onboarding flow fights the same problems:
          </p>
          <ul className="space-y-4 text-left max-w-md mx-auto mb-10">
            {problems.map((problem, index) => (
              <li key={index} className="flex items-center gap-3">
                <span className="flex-shrink-0 w-6 h-6 bg-red-100 rounded-full flex items-center justify-center">
                  <svg
                    className="w-4 h-4 text-red-500"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M6 18L18 6M6 6l12 12"
                    />
                  </svg>
                </span>
                <span className="text-[#475569] text-lg">{problem}</span>
              </li>
            ))}
          </ul>
          <p className="text-xl font-semibold text-[#0f172a]">
            Passflow fixes this without adding friction.
          </p>
        </div>
      </div>
    </section>
  );
};

export default Problem;
