function copyLink() {
    /* Copy the link into the clipboard */
    navigator.clipboard.writeText(document.getElementById("short").textContent);
    /* Tell that the link is copied */
    document.getElementById("copy").innerHTML = "Lien copié";
}

function revertCopy() {
    /* Reset the value for the copy button */
    document.getElementById("copy").innerHTML = "Copier le lien";
}

function revealPass() {
    /* Prepare a string to concatenate */
    let accssibleStr = "Accessible via le mot de passe :";
    document.getElementById("pass").innerHTML = accssibleStr.concat(" ", document.getElementById("password").value);
    /* Tell that the password is revealed */
    document.getElementById("reveal").innerHTML = "Mot de passe révélé";
}

document.getElementById("copy").addEventListener("click", copyLink); 
document.getElementById("copy").addEventListener("mouseout", revertCopy); 
document.getElementById("reveal").addEventListener("click", revealPass); 
