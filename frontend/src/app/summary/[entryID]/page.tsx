import { summaryComponents } from '@/mdx-components';
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
    const response = await fetch(`https://zircon.socialcoding.net/assets/${entryID}/Summary.txt`)
    if (response.ok) {
        const text = await response.text();
        return <MDXRemote components={summaryComponents} source={text} options={
            {
                mdxOptions: {
                    remarkPlugins: [[remarkGfm]],
                    rehypePlugins: [[rehypePrettyCode, shikiOptions]]
                }
            }
        }/>;
    }
}

export default async function RemoteMDXPage({params}:{params: Promise<{entryID: string}>}) {
    const { entryID } = await params;
    return await generateProdMarkdown(entryID);
}