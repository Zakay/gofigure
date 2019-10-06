# Examples
In gofigure, there are a number of features designed to make your life easier.

### The expansion macro
One of these features is the expansion macro. You can effectively expand roots as children of the root structure, or any other keymap.
```
# Can use both comma and space here
#            |      |
#            V      V
[root.%{root1,root2} root3]
# The statement above will prefix everything below this up until the file ends 
# or until the section is left with the "[]" operator, with root.root1., root.root2. and root3. 
# (creating a copy for every root within the expansion macro)

# Arrays
arrayKey: [
  "value",  ; comma if you want
  "value2"  # or don't
  "value3"  
]

# Objects
objectKey: {
  root3.arrayKey # Get value by identifier name
}

; This is also a comment

[] # leave the current section

rootKey: "rootkey"
```
The fig code above effectively expands the keys `root1` and `root2` as children of `root`, and `root3` as a separate root. Then it prefixes everything in that section with those keys as parents. In JSON, it looks like this
```
{
  "root": {
    "root1": {
      "key": ["value", "value2", "value3"]
    },
    "root2": {
      "key": ["value", "value2", "value3"]
    }
  },
  "root3": {
    "key": ["value", "value2", "value3"]
  },
  "rootKey": "rootkey"
}
```

### The @ selector
Now, if we want to select all children of a parent we can to so by using the `@` operator. The `@` operator is a selector which selects all predefined maps at the given level.
Say for example we wanted to add something to both root1 and root2 in the above example. We would add to the end of the file:
```
[root.@]
added: "more value"
```
As a result, both `root1` and `root2` along with any other keys (which may be) defined as children of `root`, will now contain the added value. For sake of completeness, the JSON will now look like this
```
{
  "root": {
    "root1": {
      "added": "more value",
      "key": ["value", "value2", "value3"]
    },
    "root2": {
      "added": "more value",
      "key": ["value", "value2", "value3"]
    }
  },
  "root3": {
    "key": ["value", "value2", "value3"]
  },
  "rootKey": "rootkey"
}
```

### Includes
We don't want to keep all configurations in the same file when they grow above a certain size. Enter, the %include keyword. If we have a file containing what we have below, named `firstFile.fig`
```
key: "value"
```
And we don't want to have to rewrite everything in that file. Let's just include it in the next file.
```
%include "firstFile.fig"
key2: "value"
```
Now, parsing this next file shall output the key configuration from both files, in positionally correct order. Namely if we overwrite something which would be inside the first include file in the second one, it simply gets overwritten as if they had been the same file, or merged if they are what will be an object in JSON. This is useful for concatenating larger configurations which may get used in several different places.

