import { defineConfig } from "vitepress";

// https://vitepress.dev/reference/site-config
export default defineConfig({
	title: "Bunster",
	description:
		"Compile shell scripts to native self-contained executable programs",
	themeConfig: {
		// https://vitepress.dev/reference/default-theme-config
		nav: [
			{ text: "Home", link: "/" },
			{ text: "Examples", link: "/markdown-examples" },
		],

		sidebar: [
			{
				text: "Examples",
				items: [
					{ text: "Markdown Examples", link: "/markdown-examples" },
					{ text: "Runtime API Examples", link: "/api-examples" },
				],
			},
		],

		socialLinks: [
			{ icon: "github", link: "https://github.com/yassinebenaid/bunster" },
		],
	},
	head: [
		["link", { rel: "manifest", href: "/site.webmanifest" }],
		["link", { rel: "icon", type: "image/x-icon", href: "/favicon.ico" }],
	],
});
