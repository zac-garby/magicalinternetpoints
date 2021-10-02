function login() {
    var email = document.getElementById("login-email").value
    var password = document.getElementById("login-password").value

    var form = new FormData()
    form.set("email", email)
    form.set("password", password)

    var req = new XMLHttpRequest()

    req.onreadystatechange = function() {
        if (this.readyState != 4) return;

        var resp = JSON.parse(this.responseText)

        if (this.status == 200) {
            console.log("logged in. session id =", resp)
            document.cookie = "sessionID=" + resp.session_id + "; expires=" + resp.expiry + "; path=/;"
            
            window.location.pathname = "/"
        } else {
            console.error(resp)
        }
    }

    req.open("POST", "/api/login")
    req.send(form)
}