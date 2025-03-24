# dead-pool
A Golang script that automaticaly grade a post CC Pool project, based on their sub-modules marks.

## How to use
1. git clone ssh://git@gitlab.42nice.fr:4222/42nice/dead-pool.git
2. edit main.go, to setup your API Keys. Your application must have [public, projects] scopes and [Basic Tutor, Advanced Tutor, Basic Staff, Advanced Staff] as roles
3. go run main.go <userID> <poolModuleID>

## What will happen
Based on the module id you provide, dead-pool will find the parent project.  
After that, it will check that all piscine module are validated (green in intranet), and will calculate the final_mark for the parent project. (Average of all childrens)  
If the final_mark is higher than the current parent project's mark, it will apply the mark to the latest team of the user, and to the projects_users object.

## When to use
This script is meant to be used by Captain-hook, when any pool project is validated.  
The poolModuleID is validated by dead-pool, so you can call dead-pool on any validated project if you are lazy.  
Only modules listed in pool-list.json will work.  

## How to maintain
Since 42 API is not complete, we have no way to know which pool project is related to which parent project.  
You can read and edit file pool-list.json, that is used to link projects together.  
If 42 Central update a Pool, adding, or removing a module from it, you must come to this file, and add the new projectID in the list.  