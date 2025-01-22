import { components } from '@/mdx-components';
import { MDXRemote } from 'next-mdx-remote/rsc';
import rehypePrettyCode from 'rehype-pretty-code';
import remarkGfm from 'remark-gfm';

// Disables revalidation
export const revalidate = false;
// Render from static params, allow dynamic params (run-time)
export const dynamicParams = true;
export const dynamic = 'force-static';

const shikiOptions = {
    theme: "vitesse-black",
}

async function generateProdMarkdown(entryID:string){
    try {
        const response = await fetch(`https://analysis.socialcoding.net/assets/${entryID}/Notes.md`)
        if (!response.ok) {
            return (
                <div className="flex w-full min-h-64 text-center text-5xl lg:text-6xl text-foreground">
                    <div className="m-auto">{"404: No Notes Found :("}</div>
                </div>
            )
        }
        const text = await response.text();
        return <MDXRemote components={components} source={text} options={
            {
                mdxOptions: {
                    remarkPlugins: [[remarkGfm]],
                    rehypePlugins: [[rehypePrettyCode, shikiOptions]]
                }
            }
        }/>;
    } catch {
        return (
            <div className="flex w-full min-h-64 text-center text-5xl lg:text-6xl text-foreground">
                <div className="m-auto">{"500: Server Couldn't Be Contacted :("}</div>
            </div>
        )
    }
}

export default async function RemoteMDXPage({params}:{params: Promise<{entryID: string}>}) {
    const { entryID } = await params;
    return await generateProdMarkdown(entryID);
}