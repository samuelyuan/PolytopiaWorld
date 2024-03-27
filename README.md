# PolytopiaWorld

## Introduction

This program renders The Battle of Polytopia maps using isometric projection similar to the real game. This allows you to view the map outside of the original game. 

### Controls

- W/S to move up/down
- A/D to rotate left/right
- Scroll up/down to zoom in/out

### Generated Output

To view the map, you must have a game in progress. The input filename is a .state file from the save game directory. Make sure to copy the .state file in a different folder, such as this project, because the .state file will be deleted once the game ends. The .state file can't be recovered after it's deleted.

Command format:
```
./PolytopiaWorld.exe -input=[input state filename]
```

Example:
```
./PolytopiaWorld.exe -input=00000000-0000-0000-0000-000000000000.state
```

<table>
  <tr>
    <td>Screenshot from this program</td>
    <td>Original game for reference</td>
  </tr>
  <tr>
    <td><img src="https://raw.githubusercontent.com/samuelyuan/PolytopiaWorld/master/screenshots/viewer.png" width="400" height="300"></td>
    <td><img src="https://raw.githubusercontent.com/samuelyuan/PolytopiaWorld/master/screenshots/original.jpg" width="480" height="300"></td>
  </tr>
 </table>
 

