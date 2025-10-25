document.addEventListener('DOMContentLoaded', function() {
  const rssButton = document.getElementById('rss-button');
  const rssButtonText = document.getElementById('rss-button-text');
  const authorFilter = document.getElementById('author-filter');

  if (!authorFilter && !rssButton) return;

  // Function to update RSS button
  function updateRssButton(selectedAuthor) {
    if (!rssButton || !rssButtonText) return;

    if (selectedAuthor) {
      // Update to author-specific feed
      rssButton.href = '/authors/' + encodeURIComponent(selectedAuthor) + '/feed.xml';
      rssButtonText.textContent = selectedAuthor + ' の RSS を購読';
    } else {
      // Reset to main feed
      rssButton.href = '/feed.xml';
      rssButtonText.textContent = 'RSS を購読';
    }
  }

  // If there's no author filter, just initialize the RSS button and exit
  if (!authorFilter) {
    updateRssButton('');
    return;
  }

  // Function to apply the filter
  function applyFilter(selectedAuthor) {
    const dateGroups = document.querySelectorAll('.date-group');

    dateGroups.forEach(function(dateGroup) {
      const posts = dateGroup.querySelectorAll('.post-list li');
      let visiblePostsCount = 0;

      posts.forEach(function(post) {
        const postAuthors = post.getAttribute('data-authors');

        if (!selectedAuthor || !postAuthors) {
          // Show all posts if no filter selected or post has no authors
          post.style.display = '';
          if (postAuthors || !selectedAuthor) visiblePostsCount++;
        } else {
          // Check if the selected author is in the post's authors
          const authorsList = postAuthors.split(',');
          if (authorsList.includes(selectedAuthor)) {
            post.style.display = '';
            visiblePostsCount++;
          } else {
            post.style.display = 'none';
          }
        }
      });

      // Hide the entire date group if no posts are visible
      if (visiblePostsCount === 0) {
        dateGroup.style.display = 'none';
      } else {
        dateGroup.style.display = '';
      }
    });

    // Update RSS button to match filter
    updateRssButton(selectedAuthor);
  }

  // Restore saved filter from localStorage
  const savedAuthor = localStorage.getItem('authorFilter');
  if (savedAuthor) {
    authorFilter.value = savedAuthor;
    applyFilter(savedAuthor);
  } else {
    // Initialize RSS button on page load
    updateRssButton('');
  }

  // Save filter selection and apply filter when changed
  authorFilter.addEventListener('change', function() {
    const selectedAuthor = this.value;
    localStorage.setItem('authorFilter', selectedAuthor);
    applyFilter(selectedAuthor);
  });
});
