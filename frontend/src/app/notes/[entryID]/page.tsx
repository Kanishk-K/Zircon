import { useMDXComponents } from '@/mdx-components';
import { DynamoDBClient } from '@aws-sdk/client-dynamodb';
import { GetObjectCommand, S3Client } from '@aws-sdk/client-s3';
import { GetCommand } from '@aws-sdk/lib-dynamodb';
import { MDXRemote } from 'next-mdx-remote/rsc';

function notFoundPage (){
    return (
        <div className='flex flex-row w-full justify-center items-center h-[60vh]'>
            <h1 className='text-5xl lg:text-7xl font-semibold leading-snug text-center'><span className="text-brand">404</span> - No Notes Found Here!</h1>
        </div>
    )
}

// Refresh pages after 7 days.
export const revalidate = 60 * 60 * 24 * 7;
// Render from static params, allow dynamic params (run-time)
export const dynamicParams = true;

export async function getStaticPaths() {
  return {
    paths: [],
    fallback: 'blocking',
  };
}

const credentials = {
    region: process.env.AWS_REGION as string,
    credentials: {
        accessKeyId: process.env.AWS_ACCESS_KEY_ID as string,
        secretAccessKey: process.env.AWS_SECRET_ACCESS_KEY as string,
    },
}

const s3Client = new S3Client(credentials);
const dbClient = new DynamoDBClient(credentials)

export default async function RemoteMDXPage({params}:{params: Promise<{entryID: string}>}) {
    const { entryID } = await params;
    const command = new GetCommand({
        TableName: process.env.AWS_DYNAMODB_TABLE as string,
        Key: {
            entryID: entryID
        },
    })
    const dbResponse = await dbClient.send(command);
    if (!dbResponse.Item) {
        return notFoundPage();
    }

    const s3Response = await s3Client.send(new GetObjectCommand({
        Bucket: process.env.AWS_BUCKET_NAME as string,
        Key: `${entryID}/Notes.md`,
    }));
    if (!s3Response.Body) {
        return notFoundPage();
    }
    const text = await s3Response.Body.transformToString();

    return <MDXRemote components={useMDXComponents({})} source={text} />;
}