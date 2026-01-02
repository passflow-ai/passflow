import Header from "@/components/Header";
import Hero from "@/components/Hero";
import Urgency from "@/components/Urgency";
import ValueProp from "@/components/ValueProp";
import Differentiation from "@/components/Differentiation";
import HowItWorks from "@/components/HowItWorks";
import UseCases from "@/components/UseCases";
import WhoItsFor from "@/components/WhoItsFor";
import Trust from "@/components/Trust";
import FinalCTA from "@/components/FinalCTA";
import Footer from "@/components/Footer";

export default function Home() {
  return (
    <>
      <Header />
      <main>
        <Hero />
        <Urgency />
        <ValueProp />
        <Differentiation />
        <HowItWorks />
        <UseCases />
        <WhoItsFor />
        <Trust />
        <FinalCTA />
      </main>
      <Footer />
    </>
  );
}
