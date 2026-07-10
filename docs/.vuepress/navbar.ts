import { defineNavbarConfig } from "vuepress-theme-plume";

export const navbar = defineNavbarConfig([
  { text: "Home", link: "/", icon: "mdi:home" },

  {
    text: "Getting Started",
    icon: "mdi:rocket-launch",
    items: [
      { text: "Overview", link: "/getting-started/overview", icon: "mdi:eye" },
      {
        text: "Installation",
        link: "/getting-started/installation",
        icon: "mdi:download",
      },
      {
        text: "Quick Start",
        link: "/getting-started/quick",
        icon: "mdi:flash",
      },
    ],
  },

  {
    text: "Guides",
    icon: "mdi:compass",
    items: [
      {
        text: "Enter a context",
        link: "/guides/enter-context",
        icon: "mdi:layers",
      },
      {
        text: "Switch namespaces",
        link: "/guides/namespaces",
        icon: "mdi:folder-swap",
      },
      { text: "Run one command", link: "/guides/exec", icon: "mdi:play-box" },
      {
        text: "Config & discovery",
        link: "/guides/configuration",
        icon: "mdi:file-cog",
      },
      {
        text: "Tab-completion",
        link: "/guides/completion",
        icon: "mdi:keyboard",
      },
      {
        text: "Prompt integration",
        link: "/guides/prompt",
        icon: "mdi:console-line",
      },
      {
        text: "Guard kubectl outside kush",
        link: "/guides/guard",
        icon: "mdi:shield-alert",
      },
      {
        text: "Agent Plugins",
        link: "/guides/agents",
        icon: "mdi:robot-happy",
      },
    ],
  },

  {
    text: "Understanding",
    icon: "mdi:lightbulb",
    items: [
      {
        text: "How isolation works",
        link: "/understanding/isolation",
        icon: "mdi:shield-lock",
      },
    ],
  },

  {
    text: "Reference",
    icon: "mdi:book",
    items: [
      { text: "CLI Reference", link: "/reference/cli", icon: "mdi:terminal" },
      {
        text: "Configuration",
        link: "/reference/configuration",
        icon: "mdi:file-cog",
      },
    ],
  },

  {
    text: "More",
    icon: "mdi:dots-horizontal",
    items: [
      {
        text: "Download",
        link: "https://github.com/spechtlabs/kush/releases",
        target: "_blank",
        rel: "noopener noreferrer",
        icon: "mdi:download",
      },
      {
        text: "Report an Issue",
        link: "https://github.com/spechtlabs/kush/issues/new/choose",
        target: "_blank",
        rel: "noopener noreferrer",
        icon: "mdi:bug-outline",
      },
    ],
  },
]);
