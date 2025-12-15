# Reasoning

This document serves for adding reasoning to the choices
made throughout the assignment implementation and for other observations.

## Router/web framework choice

At first, I decided to construct objects I will be working with -> OpenAPI schemas in models.go file, 
whilst trying to minimalize duplicants.
That led to unifying Responses from Reviews and Recommendations to JSONResponse.

After completing modules I continued in cmd/main.go with setting up a router using chi.
Chi I chose based on recommendation at lecture, simplicity, compatibility, ability to parse statements.

Then I moved to Handlers which contain logic for interaction between obtained user input and Storage.
I opened OpenAPI manuals for each scenario and started by receiving information. Then followed scenarios 
descried for different status codes 200, 400, 500 and handled cases. If it was invalid IDs I used library 
uuid to check them. For time, library time from lecture 4 and used custom format.

Doing Handlers before DB logic helps me to determine which functionality is needed and temporarily declare pseudocode,
that will be later written.

Then I moved to implement functions to storage.go that were marked by pseudocode. I used logic from my project from last
semester and made DB relations <br>
```
Storage {
    store Object1 (userID -> Reviews{})
    store Object2 (contentID -> Review{})
    
    store stored Object (reviewID -> Review)    
}
```
This help with fast lookup, whilst keeping Add and delete straight forward, but takes more space.
After implementing Constructor, HasUser, HasContent, Add/Delete, I moved onto recommendation algorithm and logic
with same marking logic as in handlers.

At last, recommendation algorithm. I decided to use as main metric genre and tags. Basic user tends to enjoy his comfort
pick which are mostly movies/series of same genre or with similar theme -> captured in tags. <br>

Content -> Content: for similar genres produce similar genres

Content -> User: User want to explore new film and not cycle in 3 movie loop forever