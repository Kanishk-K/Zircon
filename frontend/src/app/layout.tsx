import type { Metadata } from "next";
import { Geist } from "next/font/google";
import "./globals.css";
import Navbar from "../components/layout/navbar";
import Footer from "@/components/layout/footer";

const geist = Geist({
  variable: "--font-geist",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Lecture Analyzer Application",
  description: "Analyze lectures at the University of Minnesota.",
};

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
