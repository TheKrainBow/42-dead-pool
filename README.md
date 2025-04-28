# dead-pool
A Golang script that automaticaly grade a post CC Pool project, based on their sub-modules marks.

## How to use
1. Clone the Project
2. edit main.go, to setup your API Keys. Your application must have [public, projects] scopes and [Basic Tutor, Advanced Tutor, Basic Staff, Advanced Staff] as roles
3. go run main.go \<userID\> \<poolModuleID\>

## What will happen
Based on the module id you provide, dead-pool will find the parent project.  
After that, it will check that all piscine module are validated (green in intranet), and will calculate the final_mark for the parent project. (Average of all childrens)  
If the final_mark is higher than the current parent project's mark, it will apply the mark to the latest team of the user, and to the projects_users object.

## Usage

This script is designed to be used with services that check **42Webhooks** for validated projects.

* The `poolModuleID` variable is checked by the `dead-pool` script.
* Any invalid IDs not found in the configuration (`pool-list.json`) file will be ignored by the script.
* Alternatively, you can run this script **independently** (on its own) if you do not have services that check webhooks. 

## How to maintain
Since the relationship between parent projects and child projects are not handled by the API, we have no way to know which pool project is related to which parent project.  
You can read and edit file (`pool-list.json`), that is used to link projects together.  
If 42World update a Pool, adding, or removing a module from it, you must come to this file, and add the new projectID in the list.  
The list provided in (`pool-list.json`) is the valid for 42Nice, the 28th April 2025.  

## Conditions of Use

Copyright (c) [2025] AGOSTINI Mathieu. All Rights Reserved.

You are granted permission to use this software **only in its original, unmodified form**. 
Any modification, adaptation, redistribution, or reverse engineering of this software is strictly prohibited. 

This software is provided "AS IS", without warranty of any kind. The provider assumes no liability arising from its use.

By using this software, you signal your agreement to these conditions.

For any modification, please create a merge-request on this repo.