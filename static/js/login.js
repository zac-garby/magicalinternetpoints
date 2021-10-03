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

function register() {
    var username = document.getElementById("register-username").value
    var email = document.getElementById("register-email").value
    var password = document.getElementById("register-password").value

    var form = new FormData()
    form.set("username", username)
    form.set("email", email)
    form.set("password", password)

    post("/api/register", form, function(resp) {
        console.log("registered a new user")
        window.location.pathname = window.location.pathname
    }, function(resp) {
        switch (resp.code) {
        case "USER_ALREADY_EXISTS":
            setError("A user already exists with that username or email. Is it you?")
            break
        default:
            setError("An unexpected server error occurred. Please try again later.")
            break
        }
    })
}