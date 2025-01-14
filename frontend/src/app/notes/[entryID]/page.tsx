import { components } from '@/mdx-components';
import { DynamoDBClient } from '@aws-sdk/client-dynamodb';
import { GetObjectCommand, S3Client } from '@aws-sdk/client-s3';
import { GetCommand } from '@aws-sdk/lib-dynamodb';
import { MDXRemote } from 'next-mdx-remote/rsc';
import { redirect } from 'next/navigation';
import rehypePrettyCode from 'rehype-pretty-code';
import remarkGfm from 'remark-gfm';

// Refresh pages after 7 days.
export const revalidate = false;
// Render from static params, allow dynamic params (run-time)
export const dynamicParams = true;

const credentials = {
    region: process.env.AWS_REGION as string,
    credentials: {
        accessKeyId: process.env.AWS_ACCESS_KEY_ID as string,
        secretAccessKey: process.env.AWS_SECRET_ACCESS_KEY as string,
    },
}

const s3Client = new S3Client(credentials);
const dbClient = new DynamoDBClient(credentials)

const shikiOptions = {
    theme: {
        dark: "vitesse-black",
        light: "vitesse-light",
    }
}

async function generateProdMarkdown(entryID:string){
    const command = new GetCommand({
        TableName: process.env.AWS_DYNAMODB_TABLE as string,
        Key: {
            entryID: entryID
        },
    })
    const dbResponse = await dbClient.send(command);
    if (!dbResponse.Item || !dbResponse.Item.notesGenerated) {
        try {
            throw new Error('Entry not found in database!');
        } catch (e) {
            console.log (e);
            redirect("/")
        }
    } 

    try {
        const s3Response = await s3Client.send(new GetObjectCommand({
            Bucket: process.env.AWS_BUCKET_NAME as string,
            Key: `${entryID}/Notes.md`,
        }));
        if (!s3Response.Body) {
            try {
                throw new Error('No body in S3 response!');
            } catch (e) {
                console.log(e);
                redirect("/");
            }
        }
        const text = await s3Response.Body.transformToString();
        return <MDXRemote components={components} source={text} options={
        {
            mdxOptions: {
                remarkPlugins: [[remarkGfm]],
                rehypePlugins: [[rehypePrettyCode, shikiOptions]]
            }
        }
    }/>;
    } catch (e) {
        console.log(e);
        redirect("/");
    }
}

export default async function RemoteMDXPage({params}:{params: Promise<{entryID: string}>}) {
    const { entryID } = await params;
    return await generateProdMarkdown(entryID);
}