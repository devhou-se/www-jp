# Dev House Blog Platform

This document provides a comprehensive guide to the Dev House blogging platform. The platform uses GitHub Issues as the primary content source for blog posts, with an automated workflow that converts issues into Hugo-compatible Markdown files, optionally translates them to Japanese, and deploys the static site.

## Table of Contents

- [System Overview](#system-overview)
- [Content Creation](#content-creation)
- [Automated Workflows](#automated-workflows)
- [Translation System](#translation-system)
- [Image Management](#image-management)
- [Reactions System](#reactions-system)
- [Game Integration](#game-integration)
- [Deployment](#deployment)
- [Local Development](#local-development)
- [Troubleshooting](#troubleshooting)

## System Overview

The Dev House blog platform consists of several components:

1. **Content Creation**: Blog posts are written as GitHub Issues with the "post" label
2. **Automated Workflows**: GitHub Actions that process issues into blog posts
3. **Translation System**: Automatic translation of posts to Japanese using OpenAI
4. **Image Management**: Special handling of images for gallery and daily image
5. **Reactions System**: Integration with GitHub's reaction system for post engagement
6. **Game Integration**: Connection to a Godot-based game that uses blog content
7. **Deployment**: Automatic building and deployment to Firebase Hosting

The platform is built on [Hugo](https://gohugo.io/), a fast static site generator, with a custom theme designed for Dev House.

## Content Creation

### Creating a New Blog Post

1. Go to [new blog post](https://github.com/devhou-se/www-jp/issues/new?labels=post)
2. Add a title and content for your post
3. Submit the issue with the "post" label
4. Comment `/post` on the issue to trigger the workflow
5. Optionally, use `/post --no-translate` to skip translation

### Content Format

Blog posts follow a simple format:

```markdown
---
title: My Post Title
date: 2025-03-19T12:00:00+09:00
author: username
---

Your post content here...

![Image description](/images/my-image.jpg)
```

The header section (between `---`) can include:
- `title`: Post title
- `date`: Publication date (ISO format)
- `author`: GitHub username
- `draft`: Set to "true" to mark as draft

You can include this header in your issue, or it will be automatically generated from the issue metadata.

## Automated Workflows

### Post Creation Workflow

The main workflow that processes blog posts is defined in `.github/workflows/poster.yaml` and triggered when a comment with `/post` is made on an issue with the "post" label.

Steps:
1. Validates the comment trigger
2. Creates a branch for the post
3. Processes the issue content into a Markdown file
4. Optionally translates the content to Japanese
5. Notifies the game repository (if game integration is enabled)
6. Commits the changes and creates a PR
7. Merges the PR into the main branch
8. Updates the gallery and deploys the site

### Gallery Update Workflow

The gallery workflow (`.github/workflows/gallery.yaml`) collects images from recent posts and updates the gallery display:

1. Scans all posts for images
2. Filters images based on post age and status
3. Updates the gallery HTML with recent images
4. Creates a dedicated gallery page

### Deploy Workflow

The deploy workflow (`.github/workflows/deploy.yaml`) builds and deploys the site:

1. Processes images using the Go imager tool
2. Builds the site using Hugo
3. Deploys to Firebase Hosting
4. Builds and pushes a Docker image

## Translation System

The platform includes automatic translation of posts to Japanese using OpenAI:

1. The `translator.py` script analyzes the post content
2. If the text is primarily in English (Japanese content < 50%), it's sent to OpenAI
3. The translation is performed by a specialized OpenAI Assistant
4. The translated content replaces the original in the Markdown file

To skip translation for a post, use the comment `/post --no-translate`.

## Image Management

### Adding Images to Posts

Include images in your posts using standard Markdown syntax:

```markdown
![Image description](/images/my-image.jpg)
```

Images are automatically processed and included in the gallery if they meet certain criteria.

### Gallery System

The gallery system:
- Collects images from recent posts (less than 2 days old by default)
- Displays random images in the sidebar
- Creates a dedicated gallery page with all images
- Links each image to its source post

## Reactions System

The platform integrates with GitHub's reaction system:

1. A Python Flask server (`reactions.py`) handles reaction requests
2. Reactions from GitHub Issues are displayed on the corresponding blog posts
3. Users can react to posts directly on the blog, which updates the GitHub Issue

## Game Integration

The platform includes integration with a Godot-based game:

1. When a new post is created, a webhook notification is sent to the game repository
2. The game processes the post content to:
   - Generate character data and dialogue
   - Create voice synthesis for dialogue
   - Update the game with new character data
3. Characters in the game represent blog authors, with personality and dialogue derived from posts

For detailed information, see the [Game Integration Guide](./game-integration/README.md).

## Deployment

The site is automatically deployed when changes are merged to the main branch:

1. Hugo builds the static site
2. Firebase Hosting serves the content
3. A Docker container is also built and pushed for alternative hosting

## Local Development

To develop locally:

1. Install Hugo:
   ```shell
   go install -tags extended github.com/gohugoio/hugo@latest
   ```

2. Run the development server:
   ```shell
   cd site
   hugo server
   ```

The site will be available at http://localhost:1313 with live reloading.

## Troubleshooting

### Common Issues

- **Workflow Not Triggered**: Ensure the issue has the "post" label and a "/post" comment
- **Translation Failures**: Check the OpenAI API key and Assistant configuration
- **Image Processing Issues**: Verify image paths and formats
- **Deployment Failures**: Check Firebase configuration and permissions

### Logs and Debugging

View GitHub Actions logs for detailed information about workflow runs:

1. Go to the Actions tab in the repository
2. Select the relevant workflow
3. Check the logs for error messages and details

For local development issues, run Hugo with the `--verbose` flag for more detailed logs.