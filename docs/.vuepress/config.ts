import { docsKitPlugin } from "@spechtlabs/docs-kit";
import { viteBundler } from "@vuepress/bundler-vite";
import { defineUserConfig } from "vuepress";
import { plumeTheme } from "vuepress-theme-plume";

export default defineUserConfig({
  base: "/",
  lang: "en-US",
  title: "kush",
  description:
    "Ephemeral, isolated kube-context subshells. Prod in one terminal, dev in another, zero bleed.",

  head: [
    [
      "meta",
      {
        name: "description",
        content:
          "kush drops you into a throwaway subshell pinned to exactly one Kubernetes context, using a private kubeconfig that is deleted on exit. Auth-agnostic, with no global kubeconfig state to corrupt.",
      },
    ],
    ["link", { rel: "icon", type: "image/png", href: "/images/specht.png" }],
  ],

  bundler: viteBundler(),
  shouldPrefetch: false,

  // Exclude planning artifacts (gitignored) from the built site.
  pagePatterns: ["**/*.md", "!.vuepress", "!node_modules", "!superpowers"],

  plugins: [
    // The shared components and the terminal and cast containers, with
    // releases and contributors fetched from GitHub at build time
    docsKitPlugin({ github: { repos: ["SpechtLabs/kush"] } }),
  ],

  theme: plumeTheme({
    docsRepo: "https://github.com/spechtlabs/kush",
    docsDir: "docs",
    docsBranch: "main",

    editLink: true,
    lastUpdated: false,
    contributors: false,

    cache: "filesystem",
    search: { provider: "local" },

    sidebar: {
      "/getting-started/": [
        {
          text: "Getting Started",
          icon: "mdi:rocket-launch",
          prefix: "/getting-started/",
          items: [
            { text: "Overview", link: "overview", icon: "mdi:eye" },
            {
              text: "Installation",
              link: "installation",
              icon: "mdi:download",
            },
            {
              text: "Quick Start",
              link: "quick",
              icon: "mdi:flash",
              badge: "2 min",
            },
          ],
        },
      ],

      "/guides/": [
        {
          text: "How-to Guides",
          icon: "mdi:compass",
          prefix: "/guides/",
          items: [
            {
              text: "Contexts & Namespaces",
              icon: "mdi:kubernetes",
              link: "enter-context",
              items: [
                {
                  text: "Enter a context",
                  link: "enter-context",
                  icon: "mdi:layers",
                },
                {
                  text: "Switch namespaces",
                  link: "namespaces",
                  icon: "mdi:folder-swap",
                },
                { text: "Run one command", link: "exec", icon: "mdi:play-box" },
                { text: "OpenShift", link: "openshift", icon: "mdi:redhat" },
              ],
            },
            {
              text: "Configuration & UX",
              icon: "mdi:cog",
              link: "configuration",
              items: [
                {
                  text: "Config & discovery",
                  link: "configuration",
                  icon: "mdi:file-cog",
                },
                {
                  text: "Tab-completion",
                  link: "completion",
                  icon: "mdi:keyboard",
                },
                {
                  text: "Prompt integration",
                  link: "prompt",
                  icon: "mdi:console-line",
                },
                {
                  text: "Guard kubectl outside kush",
                  link: "guard",
                  icon: "mdi:shield-alert",
                },
              ],
            },
            {
              text: "Automation",
              icon: "mdi:robot",
              link: "agents",
              items: [
                {
                  text: "Agents Plugins",
                  link: "agents",
                  icon: "mdi:robot-happy",
                },
              ],
            },
          ],
        },
      ],

      "/understanding/": [
        {
          text: "Understanding kush",
          icon: "mdi:lightbulb",
          collapsed: false,
          prefix: "/understanding/",
          items: [
            {
              text: "How isolation works",
              link: "isolation",
              icon: "mdi:shield-lock",
            },
          ],
        },
      ],

      "/reference/": [
        {
          text: "Reference",
          icon: "mdi:book",
          collapsed: false,
          prefix: "/reference/",
          items: [
            { text: "CLI Reference", link: "cli", icon: "mdi:terminal" },
            {
              text: "Configuration",
              link: "configuration",
              icon: "mdi:file-cog",
            },
          ],
        },
      ],
    },

    markdown: {
      collapse: true,
      timeline: true,
      plot: true,
      mermaid: true,
      image: {
        figure: true,
        lazyload: true,
        mark: true,
        size: true,
      },
    },

    watermark: false,
  }),
});
