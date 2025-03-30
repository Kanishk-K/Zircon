import type { Metadata, Viewport } from "next";
import { Geist } from "next/font/google";
import "./globals.css";
import Navbar from "../components/layout/navbar";
import Footer from "@/components/layout/footer";

const geist = Geist({
  variable: "--font-geist",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Zircon",
  description: "Generate notes, summaries, and \"short-form video content\" from University of Minnesota lectures.",
};

export const viewport: Viewport = {
  themeColor: "#C59F63"
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body
        className={`${geist.className} antialiased`}
      >
        <Navbar />
        <main className="px-8 2xl:px-0">
          {children}
        </main>
        <Footer />
      </body>
    </html>
  );
}
