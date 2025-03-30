import { SERVERHOST, WEBSITE } from "../../src/util/info.js";
import { getUserJWT } from "../../src/util/jwt.js";

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

    payload = {
      entryID: msg.data.entryID,
      title: msg.data.title,
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

    async function checkExists() {
      const jwt = await getUserJWT();
      if (!jwt) {
        return false;
      }
      const response = await fetch(
        `${SERVERHOST}/exists?entryID=${payload.entryID}`,
        {
          method: "GET",
          headers: {
            Authorization: `Bearer ${jwt}`,
          },
        }
      );
      if (response.ok) {
        // Response exists, let's show this to the user
        const exists = await response.json();
        const existing_container = document.getElementById("existing-content");
        const notes_link = existing_container.querySelector("#notes > a");
        const summary_link = existing_container.querySelector("#summary > a");
        notes_link.href = `${WEBSITE}/notes/${payload.entryID}`;
        summary_link.href = `${WEBSITE}/summary/${payload.entryID}`;
        if (exists.videosAvailable) {
          for (const videoID of exists.videosAvailable) {
            const videoContainer = generateVideoAvailable(
              payload.entryID,
              videoID
            );
            existing_container.appendChild(videoContainer);
          }
        }
        existing_container.classList.remove("hidden");
      } else {
        if (response.status === 404) {
          // Entry does not exist, do nothing
          console.log("Entry does not exist");
        }
      }
    }

    async function handleProcess() {
      // Disable all form elements, remove event listeners, and show progress container
      this.disabled = true;
      Array.from(videoSelections).forEach((video) => {
        video.removeEventListener("click", videoClick);
        video.style.cursor = "not-allowed";
      });
      const progressContainer = document.getElementById("progress-container");
      progressContainer.classList.remove("hidden");
      const submitProgress = document.getElementById("to-server");
      submitProgress.classList.remove("hidden");
      submitProgress.classList.add("processing");

      const jwt = await getUserJWT();
      if (!jwt) {
        submitProgress.classList.add("error");
        return;
      }
      fetch(`${SERVERHOST}/submitJob`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${jwt}`,
        },
        body: JSON.stringify(payload),
      })
        .then(async (response) => {
          if (!response.ok) {
            submitProgress.classList.add("error");
            const err = await response.json();
            throw err;
          }
          return response.json();
        })
        .then((data) => {
          submitProgress.classList.add("success");
          if (data.videoGeneration != "SKIPPED") {
            const videoStatus = document.getElementById("video-generation");
            videoStatus.classList.add("success");
            videoStatus.classList.remove("hidden");
          }
          const existing_container =
            document.getElementById("existing-content");
          const notes_link = existing_container.querySelector("#notes > a");
          const summary_link = existing_container.querySelector("#summary > a");
          notes_link.href = `${WEBSITE}/notes/${payload.entryID}`;
          summary_link.href = `${WEBSITE}/summary/${payload.entryID}`;
          existing_container.classList.remove("hidden");
        })
        .catch((err) => {
          if (err.message) {
            console.log(`Payload: ${JSON.stringify(payload)}`);
            const progressAlert = document.getElementById("progress-alert");
            const alertMessage = progressAlert.querySelector("p");
            alertMessage.innerHTML = `<strong>Error:</strong> ${err.message}`;
            progressAlert.classList.remove("hidden");
          }
        });
    }

    const submitButton = document.getElementById("submit");
    checkExists();
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

function generateVideoAvailable(entryID, videoID) {
  const hrefLink = `${SERVERHOST}/assets/${entryID}/${videoID}.mp4`;
  const videoContainer = document.createElement("div");
  videoContainer.classList.add("content-available");
  videoContainer.innerHTML = `<svg
      xmlns="http://www.w3.org/2000/svg"
      width="32"
      height="32"
      fill="#000000"
      viewBox="0 0 256 256"
    >
      <path d="M213.66,82.34l-56-56A8,8,0,0,0,152,24H56A16,16,0,0,0,40,40v72a8,8,0,0,0,16,0V40h88V88a8,8,0,0,0,8,8h48V216h-8a8,8,0,0,0,0,16h8a16,16,0,0,0,16-16V88A8,8,0,0,0,213.66,82.34ZM160,51.31,188.69,80H160ZM155.88,145a8,8,0,0,0-8.12.22l-19.95,12.46A16,16,0,0,0,112,144H48a16,16,0,0,0-16,16v48a16,16,0,0,0,16,16h64a16,16,0,0,0,15.81-13.68l19.95,12.46A8,8,0,0,0,160,216V152A8,8,0,0,0,155.88,145ZM112,208H48V160h64v48Zm32-6.43-16-10V176.43l16-10Z"></path>
    </svg>`;
  const videoLink = document.createElement("a");
  videoLink.href = hrefLink;
  videoLink.classList.add("brand-color");
  videoLink.setAttribute("download", `${entryID}_${videoID}.mp4`);
  videoLink.textContent = `Download ${toTitleCase(videoID)} Video`;
  videoContainer.appendChild(videoLink);
  return videoContainer;
}

function toTitleCase(str) {
  return str
    .toLowerCase()
    .split("_")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}
