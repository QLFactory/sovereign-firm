import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone",
  // No COEP/COOP headers needed - Sandpack works without them
  // (WebContainers required them but we switched to Sandpack)
};

export default nextConfig;
