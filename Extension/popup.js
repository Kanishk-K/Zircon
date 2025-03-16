window.addEventListener("load", async (event) => {
  UNAUTH_CONTENT = document.getElementById("login-content");
  AUTH_CONTENT = document.getElementById("logged-in-content");
  AUTH = chrome.storage.local.get("AUTH", function (data) {
    if (data.AUTH !== undefined) {
      // User has a JWT authenticated
      const { email, expiry } = data.AUTH;
      if (expiry <= Date.now() / 1000) {
        // User has authenticated before, but that has expired.
        UNAUTH_CONTENT.classList.remove("hidden");
      } else {
        // User has a JWT and is authenticated
        EMAIL_DISPLAY = document.getElementById("email");
        EMAIL_DISPLAY.innerText = email;
        AUTH_CONTENT.classList.remove("hidden");
      }
    } else {
      // User is not authenticated
      UNAUTH_CONTENT.classList.remove("hidden");
      LOGIN_BUTTON = document.getElementById("login");
      LOGIN_BUTTON.addEventListener("click", (event) => {
        window.close();
      });
    }
  });
});
