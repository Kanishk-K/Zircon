window.addEventListener("load", function () {
  const player = this.document.querySelector("#playerScript");
  if (player === null) {
    logError("Player not found");
    return;
  } else {
    const partnerID = player.innerHTML.match(/"partnerId":(\d+)/)[1];
    const entryID = player.innerHTML.match(/entryId:"([^"]+)"/)[1];
    logSuccess(`Partner ID: ${partnerID}, Entry ID: ${entryID}`);
    const element = document.querySelector("#wrap");
    getVideoInformation(partnerID, entryID)
      .then(generateBanner)
      .then((banner) => {
        if (banner.children.length !== 0) {
          element.prepend(banner);
        }
      });
  }
});
