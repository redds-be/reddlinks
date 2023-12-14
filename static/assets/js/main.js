function copyLink() {
    /* Copy text into clipboard */
    navigator.clipboard.writeText
    ("{{.ShortenedLink}}");
    /* Set the new value for the button */
    var conf = document.getElementById("confirmation");
    conf.innerHTML = "Copied Link"
}
function revertCopy() {
    /* Reset the value for the button */
    var conf = document.getElementById("confirmation");
    conf.innerHTML = "Copy Link"
}
function revealPass() {
    var pass = document.getElementById("pass")
    pass.innerHTML = "Accessible using password: {{.Password}}"
    var rev = document.getElementById("reveal")
    rev.innerHTML = "Password revealed"
}