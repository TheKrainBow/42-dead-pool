# `dead-pool`

A Go script that automatically grades certain 42 Pool projects based on the marks obtained in their associated sub-modules.

## How to Use

1.  **Clone the Repository:**
    ```bash
    git clone https://github.com/TheKrainBow/42-dead-pool dead-pool
    cd dead-pool
    ```

2.  **Configure API Credentials:**
    * Edit the configuration section within `main.go` to add your 42 API client ID and secret.
    * Ensure the associated 42 API application has the required permissions:
        * **Scopes:** `public`, `projects`
        * **Roles:** `Basic Tutor`, `Advanced Tutor`, `Basic Staff`, `Advanced Staff`

3.  **Run the Script:**
    ```bash
    go run main.go <userID> <moduleID>
    ```
    * Replace `<userID>` with the target user's 42 User ID.
    * Replace `<moduleID>` with the ID of either the parent Pool project you want to evaluate or one of its child modules/projects.

## What It Does

1.  **Find Parent Project:** Using the provided `<moduleID>`, the script consults the `pool-list.json` configuration file to identify the corresponding parent project (Pool).
2.  **Check Child Modules:** It verifies that all required child modules/projects associated with that parent project (as defined in `pool-list.json`) are validated (i.e., marked green on the intranet) for the specified `<userID>`.
3.  **Calculate Mark:** It calculates a `final_mark` for the parent project, typically by averaging the marks of the validated child modules.
4.  **Apply Mark (If Higher):** If this calculated `final_mark` is higher than the user's current mark for the parent project, the script updates the mark on the user's latest team submission for that parent project via the API.

## Integration & Standalone Use

* **Integrated Use:** This script can be triggered by external services (e.g., systems monitoring 42 webhooks for project validations). These services would typically provide the `<userID>` and `<moduleID>` arguments when invoking the script.
* **Standalone Use:** You can also run the script manually from the command line as described in "How to Use", without needing integration with webhook services.
* **ID Mapping:** The script relies entirely on `pool-list.json` to map a given `<moduleID>` to its parent project. If no corresponding parent project is found for the ID in the file, the script will exit without taking action.

## Maintaining Project Relationships (`pool-list.json`)

**Background:**
The 42World API does not currently provide information linking parent projects (often referred to as "Pools") to their constituent child projects/modules. Because this relationship data isn't available automatically, the script relies on a manually maintained configuration file to understand these links.

**Manual Configuration:**
* The necessary project associations are defined in the `pool-list.json` file.
* You will need to **read and edit this file directly** to ensure the parent-to-child project links are accurate for your context.

**When Updates Are Required:**
* If 42World modifies the structure of a Pool (i.e., adds a new module/project or removes an existing one), you **must manually update** the `pool-list.json` file accordingly.
    * If a module is **added**, add its corresponding `projectID` to the relevant parent project's list within the file.
    * If a module is **removed**, remove its `projectID` from the list.

**File Validity:**
* The `pool-list.json` file included in this repository reflects the known project structure for the **42Nice campus** as of **April 28, 2025**.
* If you are using this script for a different campus, or if the project structure changes after this date, you will need to verify and potentially update this file yourself.

## Terms of Use

Copyright (c) 2025 AGOSTINI Mathieu. All Rights Reserved.

Permission is granted to use this software **strictly in its original, unmodified form**. Any modification, adaptation, redistribution, or reverse engineering of this software is prohibited.

This software is provided "AS IS", without warranty of any kind, express or implied. The author assumes no liability arising from its use.

By using this software, you agree to these terms.

## Contributing

While modification for personal use is not permitted under the terms above, contributions to improve the official `dead-pool` script are welcome. Please submit any proposed changes via a Merge Request on the repository:

[https://github.com/TheKrainBow/42-dead-pool](https://github.com/TheKrainBow/42-dead-pool)