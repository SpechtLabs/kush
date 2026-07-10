import { defineClientConfig } from "vuepress/client";
import VPContributorsCustom from "./components/VPContributorsCustom.vue";
import VPListCompare from "./components/VPListCompareCustom.vue";
import VPReleasesCustom from "./components/VPReleasesCustom.vue";

export default defineClientConfig({
  enhance({ app }) {
    app.component("VPContributors", VPContributorsCustom);
    app.component("VPReleases", VPReleasesCustom);
    app.component("VPListCompare", VPListCompare);
  },
});
