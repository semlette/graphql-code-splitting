function TextPost({ timestamp, text }) {
    const post = document.createElement("div");
    post.className = "TextPost";
    const timestampNode = document.createElement("p");
    timestampNode.className = "timestamp";
    timestampNode.textContent = timestamp;
    post.appendChild(timestampNode);
    const textNode = document.createElement("p");
    textNode.className = "text";
    textNode.textContent = text;
    post.appendChild(textNode);
    return post;
}

export default TextPost;
