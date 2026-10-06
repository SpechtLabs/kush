import { defineClientConfig } from "vuepress/client";

export default defineClientConfig({
  enhance() {
    // The shared components (VPContributors, VPReleases, VPListCompare,
    // Terminal, AsciinemaCast, ...) come from @spechtlabs/docs-kit, which
    // registers them itself (see config.ts).
  },
});
