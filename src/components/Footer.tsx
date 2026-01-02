import Link from "next/link";

const Footer = () => {
  return (
    <footer className="bg-[#0f172a] border-t border-white/10">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="flex flex-col md:flex-row items-center justify-between gap-6">
          {/* Logo and copyright */}
          <div className="flex flex-col md:flex-row items-center gap-4">
            <Link href="/" className="text-xl font-bold text-white">
              Passflow<span className="text-[#3b82f6]">.ai</span>
            </Link>
            <span className="text-white/40 text-sm">
              © {new Date().getFullYear()} Passflow
            </span>
          </div>

          {/* Links */}
          <div className="flex items-center gap-6 text-sm">
            <Link
              href="/privacy"
              className="text-white/60 hover:text-white transition-colors"
            >
              Privacy
            </Link>
            <Link
              href="/security"
              className="text-white/60 hover:text-white transition-colors"
            >
              Security
            </Link>
            <Link
              href="/about"
              className="text-white/60 hover:text-white transition-colors"
            >
              About
            </Link>
            <a
              href="mailto:hello@passflow.ai"
              className="text-white/60 hover:text-white transition-colors"
            >
              Contact
            </a>
          </div>

          {/* SOC 2 badge */}
          <div className="flex items-center gap-2 text-white/40 text-xs">
            <svg
              className="w-4 h-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"
              />
            </svg>
            <span>SOC 2 Type II</span>
          </div>
        </div>
      </div>
    </footer>
  );
};

export default Footer;
