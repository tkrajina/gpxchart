import subprocess

class Cmd:
	def __init__(self, output_file, description, params):
		self.output_file = output_file
		self.description = description
		self.params = params

cmds = [
	Cmd("simple.png", "Simple", ""),
	Cmd("with_srtm_elevations.png", "With SRTM elevations", "-srtm"),
	Cmd("simple.svg", "SVG output", "-s 200,100"),
	Cmd("custom_size.png", "Custom size", "-s 900,300"),
	Cmd("no_padding.png", "No padding", "-p 0,0,0,0"),
	Cmd("custom_padding.png", "Padding", "-p 50,20,20,20" ),
	Cmd("custom_font_size.png", "Custom font size", "-p 100,20,0,0 -f 10,20"),
	Cmd("custom_grid.png", "Custom grid", "-g 50,20"),
	Cmd("custom_labels.png", "Custom labels", "-l 250,20"),
	Cmd("custom_chart_padding.png", "Custom chart padding", "-cp 500,50,500,50"),
]


for cmd in cmds:
	command = f"gpxchart -o examples/{cmd.output_file} {cmd.params} test_files/zbevnica.gpx"
	subprocess.check_output(command.replace("  ", " ").split(" "))
	print(f"### {cmd.description}\n")
	print(f"`{command}`\n")
	print(f"![{cmd.description}](examples/{cmd.output_file})\n")

print("\n\n^^^^^ COPY TO README\n\n")