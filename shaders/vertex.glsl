#version 330

// vertex position and color data
in vec3 position;
in vec3 color;

// mvp matrices 
uniform mat4 model;
uniform mat4 view;
uniform mat4 projection;

// output the out_color variable to the next shader in the chain
out vec3 f_color;

void main(void) {
  gl_Position = projection * view * model * vec4(position, 1.0);

//   gl_Position = vec4(position, 150.0);

  // pass the color through unmodified
  f_color = color;
}