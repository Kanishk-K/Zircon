import Image from "next/image";
import Link from "next/link";
import { FaGithub } from "react-icons/fa6";

export default function Navbar() {
    return (
        <nav className="w-full flex flex-row justify-center sticky top-0 bg-background z-10 mb-8">
            <div className="flex flex-row w-full justify-between p-6 max-w-[1376px]">
                <Link href="/">
                    <Image className="h-10 lg:h-12 w-auto" src="/logo_wide.svg" alt="Logo" width={600} height={200} />
                </Link>
                <div className="hidden md:flex flex-row items-center gap-6">
                    <a href="https://www.kanishkkacholia.com/projects/zircon" target="_blank" className="hover:text-foreground text-neutral-400">About</a>
                    <Link href="/notes" className="hover:text-foreground text-neutral-400">Notes</Link>
                    <Link href="/health" className="hover:text-foreground text-neutral-400">System Status</Link>
                    <a href="https://github.com/Kanishk-K/Zircon" className="hover:text-foreground text-neutral-400 flex flex-row items-center gap-2">Github <FaGithub /></a>
                    <a target="_blank" href="https://chromewebstore.google.com/detail/afhhheecjhjoflgloafdgpbonoppknij?utm_source=website" className="p-2 bg-brand text-background font-medium rounded-lg transition-all duration-300 hover:scale-105 hover:shadow-md">Get Started</a>
                </div>
                <a href="https://github.com/Kanishk-K/Zircon" className="md:hidden text-brand font-medium rounded-lg flex items-center"><FaGithub size={'2em'} /></a>
            </div>
        </nav>
    )
}