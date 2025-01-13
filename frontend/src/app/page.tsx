import Image from "next/image";
import { FaArrowRight } from "react-icons/fa6";

export default function Home() {
  return (
    <div className={"flex flex-col gap-16"}>
      <div className="flex flex-col lg:flex-row items-center min-h-[60vh] gap-8">
        <div className="flex flex-col items-center lg:items-start gap-8 w-full lg:w-1/2 text-center lg:text-start">
          <h1 className="text-5xl lg:text-7xl font-semibold leading-snug">Turn Lectures into <span className="text-brand">Insights</span></h1>
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

      <h1 className="text-foreground text-5xl lg:text-7xl font-semibold leading-snug hidden lg:block text-center lg:mt-14">Powerful integrations and features <span className="text-brand">right out the box</span></h1>

      <div className="grid grid-cols-1 mt-8 lg:grid-cols-2 lg:h-[450px] xl:h-[550px] xl:px-24 gap-16">
        <div className="relative flex flex-col lg:flex-row items-center bg-none lg:bg-card lg:border-2 lg:border-cardborder lg:rounded-2xl lg:p-8 lg:items-end lg:justify-center">
          <div className="flex flex-col w-full gap-4">
            <p className="text-foreground font-semibold text-center text-4xl lg:text-3xl lg:font-normal"><span className="text-brand">Connect</span> with most UMN Resources</p>
          </div>
          <Image src="/Connections.svg" alt="UMN Integrations" className="lg:absolute lg:w-auto h-full object-contain blur-image" width={556} height={498}/>
        </div>
        <div className="relative flex flex-col lg:flex-row items-center bg-none lg:bg-card lg:border-2 lg:border-cardborder lg:rounded-2xl lg:p-8 lg:items-end lg:justify-center">
          <div className="flex flex-col w-full gap-4">
            <p className="text-foreground font-semibold text-center text-4xl lg:text-3xl lg:font-normal"><span className="text-brand">Download</span> unlimited HD lectures</p>
          </div>
          <Image src="/Download.png" alt="Downloading Feature" className="lg:absolute lg:w-auto h-full object-contain blur-image" width={556} height={498}/>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:h-[450px] xl:h-[550px] xl:px-24 gap-16 ">
        <div className="flex flex-col lg:flex-row items-center bg-none lg:bg-card lg:border-2 lg:border-cardborder lg:rounded-2xl lg:p-8 lg:justify-center">
          <div className="flex flex-col w-full gap-4 lg:w-1/2">
            <p className="text-foreground font-semibold text-4xl lg:text-3xl lg:font-normal"><span className="text-brand">Generate</span> a variety of short videos</p>
            <p>Choose from a selection including Minecraft, Subway Surfers, and many more!</p>
          </div>
          <div className="flex justify-center w-full lg:w-1/2">
            <Image src="/Mobile.png" alt="Mobile Feature" className="object-contain w-3/4" width={4320} height={4320}/>
          </div>
        </div>
      </div>
    </div>
  );
}
