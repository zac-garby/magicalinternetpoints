function post(path, form, success, failure) {
    var req = new XMLHttpRequest()

    req.onreadystatechange = function() {
        if (this.readyState != 4) return;

        var resp = JSON.parse(this.responseText)

        if (this.status == 200) {
            success(resp)
        } else {
            failure(resp)
        }
    }

    req.open("POST", path)
    req.send(form)
}

function setError(message) {
    document.getElementById("error").innerHTML = message
}