import Image from "next/image";
import { FaGithub, FaHouse, FaLinkedin } from "react-icons/fa6";

export default function Footer() {
    return (
        <footer className="flex flex-row w-full p-6 max-w-[1376px] border-brand border-t-[1px] justify-center items-center md:justify-between mt-8">
                <div className="flex flex-col items-center sm:flex-row bg-card border-cardborder border-2 rounded-md p-4">
                    <Image src={"/pfp.jpg"} alt="Profile Picture" width={2538} height={2538} className="rounded-full w-32 h-auto" />
                    <div className="flex flex-col justify-between p-4 text-center sm:text-left gap-3">
                        <span className="font-bold text-xl">Kanishk Kacholia</span>
                        <span className="hidden sm:block text-sm">Software Engineer</span>
                        <div className="flex flex-row gap-4 justify-center sm:justify-start">
                            <a href="https://github.com/Kanishk-K" className="text-brand"><FaGithub size={"1.5em"}/></a>
                            <a href="https://www.linkedin.com/in/kanishk-kacholia/" className="text-brand"><FaLinkedin size={"1.5em"}/></a>
                            <a href="https://www.kanishkkacholia.com/" className="text-brand"><FaHouse size={"1.5em"} /></a>
                        </div>
                    </div>
                </div>
                <div className="hidden md:flex flex-row items-center gap-6">
                    <a href="/notes" className="hover:text-foreground text-neutral-400">Notes</a>
                    <a href="#" className="hover:text-foreground text-neutral-400">System Status</a>
                    <a href="#" className="hover:text-foreground text-neutral-400">Github</a>
                    <a href="#" className="hover:text-foreground text-neutral-400">About</a>
                </div>
        </footer>
    );
}