const WhoItsFor = () => {
  const forTeams = [
    "Care about conversion, not vanity metrics",
    "Are losing revenue to fake users",
    "Want results in weeks, not quarters",
  ];

  return (
    <section className="section-padding bg-[#f8fafc]">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="max-w-4xl mx-auto">
          {/* Who it's for - positive */}
          <div className="bg-[#f0fdf4] rounded-2xl p-8 md:p-12 mb-6">
            <h3 className="text-3xl md:text-4xl font-bold text-[#0f172a] mb-6 text-center">
              Who Passflow is for
            </h3>

            <p className="text-lg text-[#475569] mb-6 text-center">
              Passflow is built for teams that:
            </p>

            <ul className="space-y-4 max-w-md mx-auto">
              {forTeams.map((item, index) => (
                <li key={index} className="flex items-center gap-3">
                  <span className="flex-shrink-0 w-6 h-6 bg-green-500 rounded-full flex items-center justify-center">
                    <svg
                      className="w-4 h-4 text-white"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M5 13l4 4L19 7"
                      />
                    </svg>
                  </span>
                  <span className="text-[#0f172a] text-lg">{item}</span>
                </li>
              ))}
            </ul>
          </div>

          {/* Who it's not for - filter */}
          <div className="bg-[#f9fafb] rounded-xl p-6 md:p-8 border border-[#e5e7eb]">
            <h4 className="text-lg font-semibold text-[#64748b] mb-2 text-center">
              Who it&apos;s not for
            </h4>
            <p className="text-[#64748b] text-sm text-center">
              If you&apos;re only checking compliance boxes, Passflow is probably not for you.
            </p>
          </div>
        </div>
      </div>
    </section>
  );
};

export default WhoItsFor;
