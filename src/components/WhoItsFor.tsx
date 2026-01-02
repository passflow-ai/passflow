const WhoItsFor = () => {
  const forTeams = [
    "Care about conversion, not vanity metrics",
    "Are losing revenue to fake users",
    "Want results in weeks, not quarters",
  ];

  return (
    <section className="section-padding bg-white">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="max-w-3xl mx-auto">
          <h3 className="text-3xl md:text-4xl lg:text-5xl font-bold text-[#0f172a] mb-8 text-center">
            Who Passflow is for
          </h3>

          <p className="text-lg text-[#475569] mb-6 text-center">
            Passflow is built for teams that:
          </p>

          <ul className="space-y-4 max-w-md mx-auto mb-12">
            {forTeams.map((item, index) => (
              <li key={index} className="flex items-center gap-3">
                <span className="flex-shrink-0 w-6 h-6 bg-[#3b82f6]/10 rounded-full flex items-center justify-center">
                  <svg
                    className="w-4 h-4 text-[#3b82f6]"
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

          <div className="border-t border-[#e2e8f0] pt-10">
            <h4 className="text-xl font-semibold text-[#0f172a] mb-4 text-center">
              Who it&apos;s not for
            </h4>
            <p className="text-[#475569] text-center">
              If you&apos;re only checking compliance boxes, Passflow is probably not for you.
            </p>
          </div>
        </div>
      </div>
    </section>
  );
};

export default WhoItsFor;
