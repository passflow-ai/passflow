import Header from "@/components/Header";
import Hero from "@/components/Hero";
import Urgency from "@/components/Urgency";
import HowItWorks from "@/components/HowItWorks";
import UseCases from "@/components/UseCases";
import Differentiation from "@/components/Differentiation";
import Trust from "@/components/Trust";
import OpenSource from "@/components/OpenSource";
import Pricing from "@/components/Pricing";
import FinalCTA from "@/components/FinalCTA";
import Footer from "@/components/Footer";

export default function Home() {
  return (
    <>
      <Header />
      <main>
        <Hero />
        <Urgency />
        <HowItWorks />
        <UseCases />
        <Differentiation />
        <Trust />
        <OpenSource />
        <Pricing />
        <FinalCTA />
      </main>
      <Footer />
    </>
  );
}
