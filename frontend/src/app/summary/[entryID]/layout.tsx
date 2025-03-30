export default function SummaryLayout({children}:{children: React.ReactNode}) {
    return (
        <div>
            <h1 className={"text-5xl lg:text-7xl"}>Summary</h1>
            {children}
        </div>
    );
}