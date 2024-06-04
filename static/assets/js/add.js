function copyLink() {
    /* Copy the link into the clipboard */
    navigator.clipboard.writeText(document.getElementById("short").textContent);
    /* Tell that the link is copied */
    document.getElementById("copy").innerHTML = "Copied Link";
}

function revertCopy() {
    /* Reset the value for the copy button */
    document.getElementById("copy").innerHTML = "Copy Link";
}

function revealPass() {
    /* Prepare a string to concatenate */
    let accssibleStr = "Accessible using password:";
    document.getElementById("pass").innerHTML = accssibleStr.concat(" ", document.getElementById("password").value);
    /* Tell that the password is revealed */
    document.getElementById("reveal").innerHTML = "Password revealed";
}