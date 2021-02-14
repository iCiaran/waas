// Adapted from https://gist.github.com/daltonnyx/716fc55bcc98d5cbfdc5ccc76544225e

var saveData = (function () {
        var a = document.createElement("a");
        document.body.appendChild(a);
        a.style = "display: none";
        return function (data, fileName) {
            blob = new Blob([data], {type: "text/plain"}),
            url = window.URL.createObjectURL(blob);
            a.href = url;
            a.download = fileName;
            a.click();
            window.URL.revokeObjectURL(url);
        };
}());
