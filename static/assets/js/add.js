/*
    reddlinks, a simple link shortener written in Go.
    Copyright (C) 2025 redd

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

function copyLink() {
    /* Copy the link into the clipboard */
    navigator.clipboard.writeText(document.getElementById("short").textContent);
    /* Tell that the link is copied */
    document.getElementById("copy").innerHTML = document.getElementById("locale-copied-link").value;
}

function revertCopy() {
    /* Reset the value for the copy button */
    document.getElementById("copy").innerHTML = document.getElementById("locale-copy-link").value;
}

function revealPass() {
    /* Prepare a string to concatenate */
    let accessibleStr = document.getElementById("locale-accessible-pass").value;
    document.getElementById("pass").innerHTML = accessibleStr.concat(" ", document.getElementById("password").value);
    /* Tell that the password is revealed */
    document.getElementById("reveal").innerHTML = document.getElementById("locale-password-revealed").value;
}

document.getElementById("copy").addEventListener("click", copyLink); 
document.getElementById("copy").addEventListener("mouseout", revertCopy); 
document.getElementById("reveal").addEventListener("click", revealPass); 
