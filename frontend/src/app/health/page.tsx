import { FaCircleExclamation, FaLayerGroup, FaServer } from "react-icons/fa6";

export const revalidate = 600; // 10 minutes

interface ServerInfo {
    id : string;
    started: string;
}

interface QueueInfo {
    processed: number;
}

interface HealthResponse {
    serverInfo?: ServerInfo[];
    queueInfo?: {[key:string]: QueueInfo};
}

function getUptime(started: string){
    const now = new Date();
    const diff = now.getTime() - new Date(started).getTime();
    const seconds = Math.floor(diff / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);

    return `${days} ${days != 1 ? "days" : "day"}, ${hours % 24} ${hours != 1 ? "hours" : "hour"}, ${minutes % 60} ${minutes != 1 ? "minutes" : "minute"}`;
}

function convertToTitleCase(str: string) {
    return str
        .toLowerCase()
        .split(" ")
        .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
        .join(" ");
}

function Error(message: string) {
    return (
        <div className={"bg-red-100 border-l-4 border-red-600 p-4 rounded-md"} role="alert">
            <div className={"flex flex-row items-center gap-2"}>
                <FaCircleExclamation className="text-red-600" />
                <p className="font-bold text-red-600">Error</p>
            </div>
            <p className="text-red-600 text-sm">{message}</p>
        </div>
    );
}

function ServerFunction(servers: ServerInfo[]) {
    return (
        <div>
            <h2 className={"text-4xl text-brand border-b-2 border-brand pb-2 pr-2 w-fit"}>Compute Fleet</h2>
            <div className={"flex flex-col gap-4 pt-4"}>
                {servers.map((server, index) => (
                    <div key={index+1} className={"flex flex-col bg-card border-cardborder border-2 justify-center gap-2 p-4 rounded-lg"}>
                        <div className={"flex flex-row items-center gap-2"}>
                            <FaServer className="text-brand" />
                            <p className="text-xl text-foreground">Server {index+1}</p>
                        </div>
                        <div className={"flex flex-col gap-2 mr-2"}>
                            <p className="text-sm">ID: {server.id}</p>
                            <p className="text-sm">Uptime: {getUptime(server.started)}</p>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    )
}

function QueueFunction(queues: {[key:string]: QueueInfo}) {
    return (
        <div>
            <h2 className={"text-4xl text-brand border-b-2 border-brand pb-2 pr-2 w-fit"}>Queue Health</h2>
            <div className={"flex flex-col gap-4 pt-4"}>
                {Object.entries(queues).map(([key, queue]) => (
                    <div key={key} className={"flex flex-col bg-card border-cardborder border-2 justify-center gap-2 p-4 rounded-lg"}>
                    <div className={"flex flex-row items-center gap-2"}>
                        <FaLayerGroup className="text-brand" />
                        <p className="text-xl text-foreground">{`Queue (Priority: ${convertToTitleCase(key)})`}</p>
                    </div>
                    <div className={"flex flex-col gap-2 mr-2"}>
                        <p className="text-sm">Processed: {queue.processed}</p>
                    </div>
                </div>
                ))}
            </div>
        </div>
    )
}

export default async function HealthPage() {
    try{
    const res = await fetch(`${process.env.SERVERHOST}/health`, {
        method: "GET",
        headers: {
            "Content-Type": "application/json",
        },
    });
    if (!res.ok) {
        return (
            Error("Unable to fetch server health information. Please try again later.")
        )
    }
    const data: HealthResponse = await res.json();
    return (
        <div className={"flex flex-col gap-16"}>
            {ServerFunction(data.serverInfo ? data.serverInfo.sort(
                (a: ServerInfo, b: ServerInfo) => {
                    return new Date(a.started).getTime() - new Date(b.started).getTime();
                }
            ) : [])}
            {QueueFunction(data.queueInfo ? data.queueInfo : {})}
        </div>
    )
    } catch {
        return (
            Error("Unable to fetch server health information. Please try again later.")
        )
    }
}