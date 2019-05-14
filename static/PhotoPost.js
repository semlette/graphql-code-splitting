function PhotoPost({ timestamp, photoURL }) {
    const post = document.createElement("div");
    post.className = "PhotoPost";
    const timestampNode = document.createElement("p");
    timestampNode.className = "timestamp";
    timestampNode.textContent = timestamp;
    post.appendChild(timestampNode);
    const imageNode = document.createElement("img");
    imageNode.src = photoURL;
    imageNode.className = "photo";
    post.appendChild(imageNode);
    return post;
}

export default PhotoPost;
