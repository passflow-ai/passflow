import Link from "next/link";

const Footer = () => {
  return (
    <footer className="bg-[#0f172a] border-t border-white/10">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="flex flex-col md:flex-row items-center justify-between gap-4">
          <div className="flex items-center gap-2">
            <Link href="/" className="text-xl font-bold text-white">
              Passflow<span className="text-[#3b82f6]">.ai</span>
            </Link>
            <span className="text-white/40">
              {new Date().getFullYear()}
            </span>
          </div>

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
            <a
              href="mailto:hello@passflow.ai"
              className="text-white/60 hover:text-white transition-colors"
            >
              Contact
            </a>
          </div>
        </div>
      </div>
    </footer>
  );
};

export default Footer;
