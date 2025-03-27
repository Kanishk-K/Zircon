const REDIRECT_URL = chrome.identity.getRedirectURL("callback");
import { CLIENT_ID, AUTH_SERVERHOST, DOMAIN } from "./src/util/info.js";
import { getUserJWT, parseJwt } from "./src/util/jwt.js";

window.addEventListener("load", async (event) => {
  const AUTH_CONTENT = document.getElementById("logged-in-content");
  const UNAUTH_CONTENT = document.getElementById("login-content");
  const LOGIN_BTN = document.getElementById("login");

  function updateProfile() {
    chrome.storage.local.get(["name"], (data) => {
      document.getElementById("name").innerText = data.name;
    });
  }

  const token = await getUserJWT();
  if (token) {
    updateProfile();
    AUTH_CONTENT.classList.remove("hidden");
  } else {
    LOGIN_BTN.addEventListener("click", async (event) => {
      chrome.identity.launchWebAuthFlow(
        {
          url: `${AUTH_SERVERHOST}/login?client_id=${CLIENT_ID}&response_type=code&scope=email+openid+profile&redirect_uri=${encodeURIComponent(
            REDIRECT_URL
          )}`,
          interactive: true,
        },
        async (redirectUrl) => {
          const url = new URL(redirectUrl);
          const code = url.searchParams.get("code");
          fetch(`${AUTH_SERVERHOST}/oauth2/token`, {
            method: "POST",
            headers: {
              "Content-Type": "application/x-www-form-urlencoded",
            },
            body: new URLSearchParams({
              grant_type: "authorization_code",
              code: code,
              client_id: CLIENT_ID,
              redirect_uri: REDIRECT_URL,
            }),
          })
            .then((res) => res.json())
            .then(async (data) => {
              if (!data.id_token) {
                throw new Error("No id_token in response");
              }
              const id_info = parseJwt(data.id_token);
              await chrome.cookies.set({
                url: `https://${DOMAIN}`,
                domain: DOMAIN,
                name: "id_token",
                value: data.id_token,
                secure: true,
                httpOnly: true,
                expirationDate: id_info.exp - 60,
              });
              await chrome.storage.local.set({
                refresh_token: data.refresh_token,
                name: id_info.name,
                email: id_info.email,
              });
              console.log(id_info);
              updateProfile();
              UNAUTH_CONTENT.classList.add("hidden");
              AUTH_CONTENT.classList.remove("hidden");
            });
        }
      );
    });
    UNAUTH_CONTENT.classList.remove("hidden");
  }
});
