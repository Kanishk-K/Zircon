const SERVERHOST = "https://zircon.socialcoding.net";
let payload = undefined;
chrome.runtime.onMessage.addListener((msg, sender) => {
  if (msg.action === "setData") {
    // you can use msg.data only inside this callback
    // and you can save it in a global variable to use in the code
    // that's guaranteed to run at a later point in time
    const errorMessage = document.getElementById("no-content-alert");
    const content = document.getElementById("process-content-container");
    if (msg.data === null) {
      console.log("Data is null");
      errorMessage.classList.remove("hidden");
      return;
    }
    /* Download Elements */
    const thumbnail = document.getElementById("thumbnail");
    const title = document.getElementById("title");
    const videoSelections = document.getElementsByClassName("video-item");

    thumbnail.src = msg.data.thumbnail;
    title.textContent = msg.data.title;
    jwt = msg.jwt;

    payload = {
      entryID: msg.data.entryID,
      title: msg.data.title,
      transcript: msg.data.transcript,
      backgroundVideo: "",
    };

    /*
      Service Selection
    */

    function videoClick() {
      // First make each video item not active
      Array.from(videoSelections).forEach((video) => {
        video.classList.remove("selected");
      });
      // Update the payload with the selected video
      // (if it matches the current selection set it to empty and do not activate)
      if (this.dataset.selection === payload.backgroundVideo) {
        payload.backgroundVideo = "";
      } else {
        payload.backgroundVideo = this.dataset.selection;
        this.classList.add("selected");
      }
    }

    Array.from(videoSelections).forEach((video) => {
      video.addEventListener("click", videoClick);
    });

    function handleProcess() {
      // Disable all form elements, remove event listeners, and show progress container
      this.disabled = true;
      Array.from(videoSelections).forEach((video) => {
        video.removeEventListener("click", videoClick);
        video.style.cursor = "not-allowed";
      });
      console.log(payload);
      const submitProgress = document.getElementById("to-server");
      submitProgress.classList.add("processing");

      fetch(`${SERVERHOST}/submitJob`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${jwt}`,
        },
        body: JSON.stringify(payload),
      })
        .then((response) => {
          if (!response.ok) {
            submitProgress.classList.add("error");
            return response.json().then((err) => {
              throw err;
            });
          }
          return response.json();
        })
        .then((data) => {
          submitProgress.classList.add("success");
          console.log(data);
        })
        .catch((err) => {
          if (err.message) {
            console.error("Error Message: ", err.message);
          }
        });
    }

    const submitButton = document.getElementById("submit");
    submitButton.addEventListener("click", handleProcess);
    content.classList.remove("hidden");
  }
});

document.addEventListener("DOMContentLoaded", () => {
  // Wait 1 second to see if the data is loaded
  setTimeout(() => {
    if (payload === undefined) {
      const errorMessage = document.getElementById("no-content-alert");
      errorMessage.classList.remove("hidden");
    }
  }, 1000);
});
