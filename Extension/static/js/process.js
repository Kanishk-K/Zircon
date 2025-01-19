const SERVERHOST = "https://analysis.socialcoding.net";
let payload = undefined;
let jwt = undefined;
const statusMapping = {
  0: "Not Queued/Requested",
  1: "Actively Processing",
  2: "Pending Processing",
  3: "Task Scheduled",
  4: "Aiming For Retry",
  5: "Task Archived/Failed",
  6: "Task Completed",
  7: "Task Aggregating Into Group",
};

function availableContentConstructor(category, link) {
  // Create an item like: <b>Category:</b> <a href="link">Link</a>
  const line = document.createElement("div");
  const bold = document.createElement("b");
  bold.textContent = `${category}: `;
  line.appendChild(bold);
  const anchor = document.createElement("a");
  anchor.href = link;
  anchor.textContent = link;
  line.appendChild(anchor);
  return line;
}

function availableVideoConstructor(videos, baseURL) {
  // Create a bulleted list of video links
  const holdingList = document.createElement("ul");
  for (const video of videos) {
    const listItem = document.createElement("li");
    listItem.appendChild(
      availableContentConstructor(video, `${baseURL}/${video}`)
    );
    holdingList.appendChild(listItem);
  }
  return holdingList;
}

function addExistingContent(elem, data, entryID) {
  if (data.notesGenerated) {
    elem.appendChild(
      availableContentConstructor("Notes", `${SERVERHOST}/notes/${entryID}`)
    );
  }
  if (data.summaryGenerated) {
    elem.appendChild(
      availableContentConstructor("Summary", `${SERVERHOST}/summary/${entryID}`)
    );
  }
  if (data.videosAvailable) {
    elem.appendChild(
      availableVideoConstructor(
        data.videosAvailable,
        `${SERVERHOST}/videos/${entryID}`
      )
    );
  }
}

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
    const title = document.getElementById("video-title");
    const hd_download = document.getElementById("hd-link");
    const sd_download = document.getElementById("sd-link");

    thumbnail.src = msg.data.thumbnail;
    title.textContent = msg.data.title;
    hd_download.href = msg.data.HD.url;
    sd_download.href = msg.data.SD.url;

    payload = {
      entryID: msg.data.entryID,
      transcript: msg.data.transcript,
      notes: false,
      summarize: false,
      backgroundVideo: "",
    };

    jwt = msg.jwt;

    // Request any current, already processed, data from the server.
    fetch(`${SERVERHOST}/existing`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: "Bearer " + msg.jwt,
      },
      body: JSON.stringify({ entryID: msg.data.entryID }),
    })
      .then((response) => response.json())
      .then((data) => {
        existingCard = document.getElementById("existing-content");
        existingCard.classList.remove("hidden");
        addExistingContent(existingCard, data, msg.data.entryID);
      })
      .catch((error) => {
        // Errors don't matter, just assume there is no data.
        console.error("Error:", error);
      });

    content.classList.remove("hidden");
  }
});

function formLogic() {
  // Add event listeners to the form elements to modify the payload
  notes_button = document.getElementById("notes");
  summary_button = document.getElementById("summary");
  video_dropdown = document.getElementById("video");

  notes_button.addEventListener("click", () => {
    payload.notes = notes_button.checked;
  });
  summary_button.addEventListener("click", () => {
    payload.summarize = summary_button.checked;
  });
  video_dropdown.addEventListener("change", () => {
    payload.backgroundVideo = video_dropdown.value;
  });
}

function updateIndividualStatus(statusElement, category, status) {
  if (status == 0 || status == 5) {
    statusElement.innerHTML = `<b>${category}: </b> ${statusMapping[status]}`;
    statusElement.classList.add("error");
  } else if (status == 6) {
    statusElement.innerHTML = `<b>${category}: </b> ${statusMapping[status]}`;
    statusElement.classList.add("success");
  } else {
    statusElement.innerHTML = `<b>${category}: </b> ${statusMapping[status]}`;
    statusElement.classList.remove("error");
  }
}

function updateStatus(statusData) {
  const notesStatus = document.getElementById("notes-status");
  const summaryStatus = document.getElementById("summary-status");
  const videoStatus = document.getElementById("video-status");

  updateIndividualStatus(notesStatus, "Notes", statusData.notesStatus);
  updateIndividualStatus(summaryStatus, "Summary", statusData.summarizeStatus);
  updateIndividualStatus(videoStatus, "Video", statusData.videoStatus);

  // If all values are 0, 5, or 6 then the job is done. Stop checking.
  for (const status of Object.values(statusData)) {
    if (status != 0 && status != 5 && status != 6) {
      return false;
    }
  }
  return true;
}

function checkJobStatusPeriodically(intervalID) {
  // Check the status of the job every second
  fetch(`${SERVERHOST}/status`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: "Bearer " + jwt,
    },
    body: JSON.stringify({ entryID: payload.entryID }),
  })
    .then((response) => response.json())
    .then((data) => {
      shouldClear = updateStatus(data);
      if (shouldClear) {
        clearInterval(intervalID);
      }
    })
    .catch((error) => {
      clearInterval(intervalID);
      console.error("Error:", error);
    });
}

function formSubmissionLogic() {
  submitButton = document.getElementById("submit");

  serverStatusCard = document.getElementById("server-status");
  submitResponse = document.getElementById("job-status");

  /* Business Logic */
  submitButton.addEventListener("click", () => {
    submitButton.disabled = true;
    serverStatusCard.classList.remove("hidden");
    // Send the payload to the server
    fetch(`${SERVERHOST}/process`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: "Bearer " + jwt,
      },
      body: JSON.stringify(payload),
    })
      .then((response) => {
        if (response.ok) {
          return response.json();
        } else {
          if (response.status === 401) {
            throw new Error("Current login unauthorized please retry!");
          } else {
            throw new Error("Server error, please try again later.");
          }
        }
      })
      .then((data) => {
        submitResponse.textContent = data.message;
        submitResponse.classList.add("success");
        intervalID = setInterval(() => {
          checkJobStatusPeriodically(intervalID);
        }, 1000);
      })
      .catch((error) => {
        submitResponse.textContent = error;
        submitResponse.classList.add("error");
      });
  });
}

document.addEventListener("DOMContentLoaded", () => {
  formLogic();
  formSubmissionLogic();
});
