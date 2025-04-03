window.addEventListener("load", function () {
  const gallery = this.document.querySelector("#galleryGrid");
  if (gallery === null) {
    logError("Gallery not found");
    return;
  } else {
    let itemList = gallery.querySelectorAll(
      ".galleryItem:not([data-is-processed])"
    );
    processItems(itemList);
    const loadMoreButton = document.querySelector(".endless-scroll-more a");
    if (loadMoreButton !== null) {
      loadMoreButton.addEventListener("click", function () {
        setTimeout(function () {
          itemList = gallery.querySelectorAll(
            ".galleryItem:not([data-is-processed])"
          );
          processItems(itemList);
        }, 1000);
      });
    }
  }
});

async function processItems(items) {
  await Promise.all(Array.from(items).map(processItem));
}

async function processItem(element) {
  // Mark the element as processed
  element.setAttribute("data-is-processed", "true");
  // If it works, add a download button above the video
  const contentImage = element.querySelector("img");
  // The Partner ID is stored in the url of the image with regex \/p\/(\d+)\/sp\/
  const partnerID = contentImage.src.match(/\/p\/(\d+)/)[1];
  // The Entry ID is stored in the url of the image with regex \/entry_id\/([^\/]+)
  const entryID = contentImage.src.match(/\/entry_id\/([^\/]+)/)[1];
  await getVideoInformation(partnerID, entryID)
    .then(generateBanner)
    .then((banner) => {
      // add an event listener to the processButton to send to the server
      if (banner.children.length !== 0) {
        element.prepend(banner);
      }
    });
}
