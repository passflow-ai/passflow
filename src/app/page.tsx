import Header from "@/components/Header";
import Hero from "@/components/Hero";
import Problem from "@/components/Problem";
import ValueProp from "@/components/ValueProp";
import HowItWorks from "@/components/HowItWorks";
import UseCases from "@/components/UseCases";
import WhyPassflow from "@/components/WhyPassflow";
import Trust from "@/components/Trust";
import FinalCTA from "@/components/FinalCTA";
import Footer from "@/components/Footer";

export default function Home() {
  return (
    <>
      <Header />
      <main>
        <Hero />
        <Problem />
        <ValueProp />
        <HowItWorks />
        <UseCases />
        <WhyPassflow />
        <Trust />
        <FinalCTA />
      </main>
      <Footer />
    </>
  );
}
