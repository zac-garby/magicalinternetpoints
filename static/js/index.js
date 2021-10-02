function onload() {
    post("/api/get-user", null, function(user) {
        document.getElementById("username").innerHTML = user.username
        document.getElementById("home").classList.remove("hidden")
    }, function(resp) {
        console.error(resp)
    })
}