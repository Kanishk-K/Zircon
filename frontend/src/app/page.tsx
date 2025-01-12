import Image from "next/image";
import { FaArrowRight } from "react-icons/fa6";

export default function Home() {
  return (
    <div className={"flex flex-col"}>
      <div className="flex flex-col lg:flex-row items-center min-h-[60vh] gap-8">
        <div className="flex flex-col items-center lg:items-start gap-8 w-full lg:w-1/2 text-center lg:text-start">
          <h1 className="text-white">Turn Lectures into <span className="text-brand">Insights</span></h1>
          <p>Analyze, summarize, and engage with your lectures like never before. Generate notes, create video content, and download content for offline use!</p>
          <div className="flex flex-row items-center gap-4">
            <a href="#" className="border-2 border-brand text-brand rounded-lg p-2 md:text-xl">Learn More</a>
            <a href="#" className="p-2 bg-brand text-background rounded-lg md:text-xl flex flex-row items-center gap-2">Get Started <FaArrowRight /></a>
          </div>
        </div>
        <div className="flex flex-col items-center lg:items-end w-full lg:w-1/2">
          <Image src="/hero.png" alt="Extension being used" className="w-full h-auto" width={2160} height={2160}/>
        </div>
      </div>
    </div>
  );
}
