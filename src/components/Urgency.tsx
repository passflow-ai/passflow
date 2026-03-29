const Urgency = () => {
  return (
    <section className="py-12 md:py-16 bg-gradient-to-r from-[#fef2f2] to-[#fff7ed]">
      <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex flex-col md:flex-row items-start gap-6">
          {/* Alert icon */}
          <div className="flex-shrink-0 w-12 h-12 bg-red-100 rounded-xl flex items-center justify-center">
            <svg
              className="w-6 h-6 text-red-500"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"
              />
            </svg>
          </div>

          <div>
            <h2 className="text-2xl md:text-3xl lg:text-4xl font-bold text-[#0f172a] mb-4">
              Your team is spending hours on tasks AI can handle in seconds.
            </h2>
            <p className="text-[#475569] mb-3">
              Manual data entry, report generation, ticket routing, follow-up emails — repetitive work piles up and slows down the teams that should be driving growth.
            </p>
            <p className="text-lg font-semibold text-[#0f172a]">
              Passflow deploys AI agents that handle it automatically, 24/7.
            </p>
          </div>
        </div>
      </div>
    </section>
  );
};

export default Urgency;
