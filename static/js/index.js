function onload() {
    post("/api/get-user", null, function(user) {
        document.getElementById("username").innerHTML = user.username
        document.getElementById("home").classList.remove("hidden")
    }, function(resp) {
        switch (resp.code) {
        case "NO_SESSION":
            window.location.pathname = "/login"
            break
        case "SESSION_EXPIRED":
            setError("Your session has expired. Please log in again.")
            window.location.pathname = "/login"
            break
        default:
            setError("An unexpected server error occurred. Please try again later.")
            break
        }
    })
}