const SERVERHOST = "http://localhost:8080";
let payload = undefined;
let jwt = undefined;
const statusMapping = {
  0: ["generation requested", "REQUEST"],
  1: ["generation currently processing", "QUEUE"],
  2: ["generation awaiting processing", "QUEUE"],
  // 3: "Task Scheduled", // Not used
  // 4: "Aiming For Retry", // Not used
  5: ["generation failed", "ERROR"],
  6: ["generation successful", "SUCCESS"],
  // 7: "Task Aggregating Into Group", // Not used
};

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
      transcript: msg.data.transcript,
      notes: false,
      summarize: false,
      backgroundVideo: "",
    };

    /*
      Service Selection
    */
    notesCheckbox = document.getElementById("notes");
    summaryCheckbox = document.getElementById("summary");
    notesCheckbox.addEventListener("change", function () {
      payload.notes = this.checked;
    });
    summaryCheckbox.addEventListener("change", function () {
      payload.summarize = this.checked;
    });

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

    jwt = msg.jwt;

    // Request any current, already processed, data from the server.
    fetch(`${SERVERHOST}/existing?entryID=${msg.data.entryID}`, {
      method: "GET",
    })
      .then((response) => {
        if (response.ok) {
          response.json();
        } else {
          throw new Error("Failed to fetch existing data");
        }
      })
      .then((data) => {
        const existingContentContainer =
          document.getElementById("existing-content");
        const existingContentLink = existingContentContainer.querySelector("a");
        existingContentLink.href = `https://www.notes.socialcoding.net/${msg.data.entryID}`;
        existingContentContainer.classList.remove("hidden");
      })
      .catch((error) => {
        // Errors don't matter, just assume there is no data.
        console.error("Error:", error);
      });

    function updateProcess(element, status, message) {
      const messageSection = element.children[2];
      messageSection.textContent = message;
      if (status === "REQUEST") {
        element.classList.add("requested");
      } else if (status === "QUEUE") {
        element.classList.add("queue");
      } else if (status === "SUCCESS") {
        element.classList.add("success");
      } else if (status === "ERROR") {
        element.classList.add("error");
      }
    }

    function handleProcess() {
      // Disable all form elements, remove event listeners, and show progress container
      this.disabled = true;
      Array.from(videoSelections).forEach((video) => {
        video.removeEventListener("click", videoClick);
        video.style.cursor = "not-allowed";
      });
      notesCheckbox.disabled = true;
      summaryCheckbox.disabled = true;
      console.log(payload);
      const submitProgress = document.getElementById("to-server");
      const notesProgress = document.getElementById("notes-gen");
      const summaryProgress = document.getElementById("summary-gen");
      const videoProgress = document.getElementById("video-gen");

      updateProcess(submitProgress, "REQUEST", "Job sending to server");
      fetch(`${SERVERHOST}/process`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${jwt}`,
        },
        body: JSON.stringify(payload),
      })
        .then((response) => response.json())
        .then((data) => {
          updateProcess(submitProgress, "SUCCESS", "Job sent to server!");
          // Initialize the statuses of the other processes
          if (payload.notes) {
            updateProcess(
              notesProgress,
              "REQUEST",
              "Notes generation requested"
            );
          }
          if (payload.summarize) {
            updateProcess(
              summaryProgress,
              "REQUEST",
              "Summary generation requested"
            );
          }
          if (payload.backgroundVideo !== "") {
            updateProcess(
              videoProgress,
              "REQUEST",
              "Video generation requested"
            );
          }
          // Poll the server for the status of the processes
          const interval = setInterval(() => {
            fetch(`${SERVERHOST}/status`, {
              method: "POST",
              headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${jwt}`,
              },
              body: JSON.stringify({ entryID: payload.entryID }),
            })
              .then((response) => response.json())
              .then((data) => {
                if (payload.notes && data.notesStatus !== undefined) {
                  updateProcess(
                    notesProgress,
                    statusMapping[data.notesStatus][1],
                    `Notes ${statusMapping[data.notesStatus][0]}`
                  );
                }
                if (payload.summarize && data.summarizeStatus !== undefined) {
                  updateProcess(
                    summaryProgress,
                    statusMapping[data.summarizeStatus][1],
                    `Summary ${statusMapping[data.summarizeStatus][0]}`
                  );
                }
                if (
                  payload.backgroundVideo !== "" &&
                  data.videoStatus !== undefined
                ) {
                  updateProcess(
                    videoProgress,
                    statusMapping[data.videoStatus][1],
                    `Video ${statusMapping[data.videoStatus][0]}`
                  );
                }
                if (
                  // If all processes are done, failed, or not requested then stop polling
                  (!payload.notes ||
                    data.notesStatus === 6 ||
                    data.notesStatus === 5) &&
                  (!payload.summarize ||
                    data.summarizeStatus === 6 ||
                    data.summarizeStatus === 5) &&
                  (payload.backgroundVideo === "" ||
                    data.videoStatus === 6 ||
                    data.videoStatus === 5)
                ) {
                  clearInterval(interval);
                  console.log("All processes done!");
                }
              })
              .catch((error) => {
                console.error("Error:", error);
                updateProcess(
                  submitProgress,
                  "ERROR",
                  "Job failed to respond with status!"
                );
                clearInterval(interval);
              });
          }, 5000);
        })
        .catch((error) => {
          console.error("Error:", error);
          updateProcess(
            submitProgress,
            "ERROR",
            "Job failed to send to server!"
          );
        });

      const progressContainer = document.getElementById("progress-container");
      progressContainer.classList.remove("hidden");
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
