# Contributing to mongoq

This is a short guide on how to contribute to the project.

## Submitting a pull request

If you find a bug that you'd like to fix, or a new feature that you'd like to implement then please submit a pull request via GitHub.


Fork the Repository:
1. Visit [mongoq repository](https://github.com/kashifkhan0771/mongoq)
2. Click the "Fork" button to create your own fork
3. Clone your fork locally:

    # Using SSH (recommended)
    git clone git@github.com:<your-username>/mongoq.git
    # Or using HTTPS
    git clone https://github.com/<your-username>/mongoq.git

    cd mongoq
Make a branch to add your new feature

    git checkout -b my-new-feature main

And get hacking.

Make sure you

* Add documentation for a new feature
* Add unit tests for a new feature
* rebase to develop `git rebase main`

When you are done with that

    git push origin my-new-feature

Your patch will get reviewed, and you might get asked to fix some stuff.

If so, then make the changes in the same branch, squash the commits, rebase it to develop then push it to GitHub with `--force`.
