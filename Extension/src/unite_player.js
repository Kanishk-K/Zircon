window.addEventListener("load", function () {
  const scripts = this.document.querySelectorAll("script");
  let partnerID = null;
  scripts.forEach((script) => {
    if (script.src && script.src.startsWith("https://cdnapisec.kaltura.com")) {
      const match = script.src.match(/\/p\/(\d+)\//);
      if (match) {
        partnerID = match[1];
      }
    }
  });
  if (partnerID === null) {
    logError("Partner ID not found");
    return;
  }
  const entryID = this.document.URL.match(/entry_id=([^&]+)/)[1];
  logSuccess(`Partner ID: ${partnerID}, Entry ID: ${entryID}`);

  const zirconContainer = this.document.createElement("div");
  zirconContainer.id = "zircon-container";

  const mainVideoContainer = this.document.createElement("div");
  zirconContainer.appendChild(mainVideoContainer);

  const element = this.document.querySelector("p");
  element.insertAdjacentElement("beforebegin", zirconContainer);

  getVideoInformation(partnerID, entryID)
    .then((data) => {
      // add an image to the mainVideoContainer
      const mainVideoImage = this.document.createElement("img");
      mainVideoImage.src = data.thumbnail;
      mainVideoContainer.appendChild(mainVideoImage);

      return data;
    })
    .then(generateBanner)
    .then((banner) => {
      mainVideoContainer.appendChild(banner);
    });

  getChildVideos(partnerID, entryID).then((data) => {
    data.forEach((child) => {
      getVideoInformation(partnerID, child.id)
        .then(generateUniteCard)
        .then((card) => {
          zirconContainer.appendChild(card);
        });
    });
  });
});
