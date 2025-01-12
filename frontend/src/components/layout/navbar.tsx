import Image from "next/image";
import { FaGithub } from "react-icons/fa6";

export default function Navbar() {
    return (
        <nav className="flex flex-row w-full justify-between p-6 max-w-[1376px] mb-8">
            <Image className="h-10 w-auto" src="/vercel.svg" alt="Logo" width={173} height={150} />
            <div className="hidden md:flex flex-row items-center gap-6">
                <a href="#" className="hover:text-white">About</a>
                <a href="#" className="hover:text-white">Notes</a>
                <a href="#" className="hover:text-white">System Status</a>
                <a href="#" className="hover:text-white flex flex-row items-center gap-2">Github <FaGithub /></a>
                <a href="#" className="p-2 bg-brand text-background font-medium rounded-lg transition-all duration-300 hover:scale-105 hover:shadow-md">Get Started</a>
            </div>
            <a href="#" className="md:hidden text-brand font-medium rounded-lg flex items-center"><FaGithub size={'2em'} /></a>
        </nav>
    )
}