const imports = {
    "TextPost.js": undefined,
    "PhotoPost.js": undefined
};

async function importModule(path) {
    if (imports[path]) {
        return imports[path];
    }
    const mod = await import(path);
    return mod.default;
}

async function getData() {
    const query = `
        query {
            posts {
                __typename,
                timestamp,
                ... Post_PhotoPost @push(module: "PhotoPost.js"),
                ... Post_TextPost @push(module: "TextPost.js"),
            }
        }

        fragment Post_PhotoPost on PhotoPost {
            photo_url,
        }
        
        fragment Post_TextPost on TextPost {
            text,
        }
    `;

    const response = await fetch("graphql", {
        method: "POST",
        body: JSON.stringify({
            query
        })
    });
    const json = await response.json();
    const posts = json.data.posts;

    for (const post of posts) {
        let postNode;
        switch (post.__typename) {
            case "TextPost": {
                const TextPost = await importModule("./TextPost.js");
                postNode = TextPost(post);
                break;
            }
            case "PhotoPost": {
                const PhotoPost = await importModule("./PhotoPost.js");
                postNode = PhotoPost(post);
                break;
            }
            default:
                throw new Error("unknown post type");
        }
        timeline.appendChild(postNode);
    }
}
