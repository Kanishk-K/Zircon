import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  /* config options here */
  pageExtensions: ['js', 'jsx', 'ts', 'tsx'],
};

 
// Merge MDX config with Next.js config
export default nextConfig