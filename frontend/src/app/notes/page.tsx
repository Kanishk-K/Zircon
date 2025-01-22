import { FaDownload } from "react-icons/fa6";
import { components } from '@/mdx-components';
import { MDXRemote } from 'next-mdx-remote/rsc';
import rehypePrettyCode from 'rehype-pretty-code';
import remarkGfm from 'remark-gfm';
import { readFileSync } from 'fs';
import path from "path";
import Link from "next/link";

const shikiOptions = {
    theme: "vitesse-black",
}

export default function ExamplePage() {
    const text = readFileSync(path.join(process.cwd(), "public/notes/example.md"), "utf-8");
    return (
        <div>
            <div className="flex flex-col mb-2 border-brand border-b-[1px] pb-10 gap-4">
                <h1 className="text-5xl lg:text-7xl">Example Notes</h1>
                <p>{"These notes were generated from a recording of Daniel Kluver's"} <a target="_blank" href="https://umn.lol/class/CSCI1913" className="text-brand hover:underline">CSCI 1913</a> {"class held on November 3rd, 2021."}</p>
                <Link href="/notes/example.md" className="inline-flex items-center p-2 bg-brand text-background rounded-lg md:text-xl w-fit" download={true}>
                    Download Raw Markdown <FaDownload className="ml-2" />
                </Link>
            </div>
            <MDXRemote components={components} source={text} options={{
                mdxOptions: {
                    remarkPlugins: [[remarkGfm]],
                    rehypePlugins: [[rehypePrettyCode, shikiOptions]]
            }}}/>
        </div>
    )
}