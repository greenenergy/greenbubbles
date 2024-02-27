# TeaTree

This is a tree "bubble" designed to be compatible with charm.sh bubbletea packages.

Operations:
- Add items to main tree
- Add items to items
- When adding an item, you can specify the icon.
    - Perhaps when doing a View() operation, the item can call an interface function to get
    its icon. This would allow clients to specify their own state icons. Should there be
    some animation support? Or is that crazy?
- Items can be opened or closed if they have children
- There should be help, though actually I guess what shows up in the help should be up to the client application. But some standard functions should exist:
    - Select (return) -- called when the user hits return on a field. Used for picking something from a hierarchy. Should the "Select" function be opt-in or opt-out? Should it do something by default, or should it do something only if a user has configured it to?
    - Open/Close 
    - Cursor Up/Down - move the selection:
        - up (previous sibling, or if at the parent, the previous sibling of the parent)
        - down (next sibling, or if at the end of the tree, we need to return to the caller that the user has tried to select down from us)
    - Cursor Right/Left - actually I won't capture these, so they could be executed by the calling application

## To Do

When rendering a tree item, it should be rendered and then if there are any children, they should be rendered. The item needs to pass the amount of lines still remaining in the view. Somehow I need to know how many lines the children used.

Is it possible to count the lines in a returned view string? If so, then I can just use that.


