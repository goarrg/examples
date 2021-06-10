# Examples
Currently this repo holds examples on component development for goarrg.com.
There are OpenGL and Vulkan hello world examples, a Vulkan fallback to OpenGL example, and an audio example.

Currently the examples support running on Ubuntu or Windows, 386 or amd64. However the Vulkan examples require amd64.
## Dependencies
### Global
| OS | Dependencies |
| -- | -- |
| Ubuntu | sudo apt-get install build-essential cmake libxext-dev libpulse-dev |
| Windows | mingw-w64, cmake |
### Graphics API Specific
| OS | Folder Prefix | Dependencies |
| -- | -- | -- |
| Ubuntu | gl | sudo apt-get install libglu1-mesa-dev mesa-common-dev |
| Ubuntu_amd64 | vk | Vulkan SDK |
| Windows_amd64 | vk | Vulkan SDK |

**For `vkgl` prefix, you need both vk and gl dependencies.**

## Setup
Once all the dependencies are installed, assuming the current directory is the examples repo, you just need to:
<pre><code>go run goarrg.com/cmd/goarrg install sdl2 -vv</code></pre>

### Vulkan
To run the Vulkan examples, after installing the SDK, assuming the current directory is the examples repo, you also need to run:
<pre><code>go run goarrg.com/cmd/goarrg install vulkan -vv</code></pre>

## Running
To run the examples, excluding the `shared` folder, cd to the folder and run:
<pre><code>go run goarrg.com/cmd/goarrg run -vv</code></pre>
