const SERVERHOST = "http://localhost:8080";

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
    hd_download.href = msg.data.download.url;
    sd_download.href = msg.data.process.url;

    /* Processing Form */
    const notesButton = document.getElementById("notes");
    const summarizeButton = document.getElementById("summary");
    const brainrotButton = document.getElementById("brainrot");
    const videoDropdown = document.getElementById("video");
    const submitButton = document.getElementById("submit");
    let videoActive = "none";
    let buttonActive = 0;
    function updateVideoActive(id) {
      if (videoActive !== "none") {
        document.getElementById(videoActive).classList.remove("active");
      }
      if (id !== "none") {
        const videoSelection = document.getElementById(id);
        videoSelection.classList.add("active");
      }
    }
    notesButton.addEventListener("change", () => {
      if (notesButton.checked) {
        buttonActive++;
        submitButton.disabled = false;
      } else {
        buttonActive--;
        if (buttonActive === 0) {
          submitButton.disabled = true;
        }
      }
    });
    summarizeButton.addEventListener("change", () => {
      if (summarizeButton.checked) {
        buttonActive++;
        brainrotButton.disabled = false;
        submitButton.disabled = false;
      } else {
        buttonActive--;
        if (buttonActive === 0) {
          submitButton.disabled = true;
        }
        brainrotButton.disabled = true;
        brainrotButton.checked = false;
        videoDropdown.disabled = true;
        videoDropdown.selectedIndex = 0;
        updateVideoActive("none");
        videoActive = "none";
      }
    });
    brainrotButton.addEventListener("change", () => {
      if (brainrotButton.checked) {
        videoDropdown.disabled = false;
      } else {
        videoDropdown.disabled = true;
        videoDropdown.selectedIndex = 0;
        updateVideoActive("none");
        videoActive = "none";
      }
    });
    videoDropdown.addEventListener("change", () => {
      updateVideoActive(videoDropdown.value);
      videoActive = videoDropdown.value;
    });

    /* Processing Timeline */
    timeline = document.getElementById("process-ongoing");
    toServer = document.getElementById("process-to-server");

    submitButton.onclick = () => {
      // Harvest the data from the form
      const submissionData = {
        download: msg.data.download.url,
        title: msg.data.title,
        notes: notesButton.checked,
        summarize: summarizeButton.checked,
        brainrot: brainrotButton.checked,
        video: videoDropdown.value,
      };
      // Disable buttons so user can't submit again
      notesButton.disabled = true;
      summarizeButton.disabled = true;
      brainrotButton.disabled = true;
      videoDropdown.disabled = true;
      submitButton.disabled = true;
      // Show to the user
      toServer.innerText = "Sending to server...";
      timeline.classList.remove("hidden");
      // Send data to the server for processing (SERVERHOST/process)
      fetch(`${SERVERHOST}/process`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(submissionData),
      })
        .then((response) => {
          if (!response.ok) {
            throw new Error(
              `Communication to server failed! status: ${response.status}`
            );
          }
          return response.json();
        })
        .then((data) => {
          toServer.classList.add("success");
          toServer.innerText = data.message;
        })
        .catch((error) => {
          console.log(error);
          toServer.classList.add("error");
          toServer.innerText = error.message;
        });
      content.classList.remove("hidden");
    };

    content.classList.remove("hidden");
  }
});
