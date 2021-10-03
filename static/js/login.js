function login() {
    var email = document.getElementById("login-email").value
    var password = document.getElementById("login-password").value

    var form = new FormData()
    form.set("email", email)
    form.set("password", password)

    post("/api/login", form, function(resp) {
        document.cookie = "session=" + resp.session_id + "; expires=" + resp.expiry + "; path=/;"

        window.location.pathname = "/"
    }, function(resp) {
        switch (resp.code) {
        case "USER_DOES_NOT_EXIST":
            setError("No user with that email exists.")
            break
        case "INVALID_PASSWORD":
            setError("Your password is incorrect.")
            break
        default:
            setError("An unexpected server error occurred. Please try again later.")
            break
        }
    })
}
