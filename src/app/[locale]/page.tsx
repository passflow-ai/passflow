import Header from "@/components/Header";
import Hero from "@/components/Hero";
import TrustBand from "@/components/TrustBand";
import Problem from "@/components/Problem";
import ValueProp from "@/components/ValueProp";
import GitHub from "@/components/GitHub";
import OpenVsEnterprise from "@/components/OpenVsEnterprise";
import FinalCTA from "@/components/FinalCTA";
import Footer from "@/components/Footer";

export default function Home() {
  return (
    <>
      <Header />
      <main>
        <Hero />
        <TrustBand />
        <Problem />
        <ValueProp />
        <GitHub />
        <OpenVsEnterprise />
        <FinalCTA />
      </main>
      <Footer />
    </>
  );
}
