// pgbook.dev content API worker. Serves site/ assets on the /api/*,
// /install.sh, and /downloads/* routes; the landing page stays on the
// separate worker bound to the domain root.
export default {
  async fetch(request, env) {
    const url = new URL(request.url);
    const path = url.pathname;

    // Static hosts can't serve a file and a directory with the same
    // name, so the topic index lives at topics.json.
    if (path === "/api/topics" || path === "/api/topics/") {
      url.pathname = "/api/topics.json";
    }

    const resp = await env.ASSETS.fetch(new Request(url, request));

    const withType = (type) => {
      const r = new Response(resp.body, resp);
      r.headers.set("content-type", type);
      return r;
    };
    if (resp.ok && path.startsWith("/api/")) {
      const r = withType("application/json; charset=utf-8");
      r.headers.set("access-control-allow-origin", "*");
      return r;
    }
    if (resp.ok && path === "/install.sh") {
      return withType("text/x-sh; charset=utf-8");
    }
    if (resp.ok && path.startsWith("/downloads/") && path.endsWith(".pdf")) {
      return withType("application/pdf");
    }
    return resp;
  },
};
