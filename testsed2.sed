# Initial substitution
s/o/0/g

# Print the pattern space before deletion
x  # Swap pattern and hold space
p  # Print the hold space (original pattern space)
x  # Swap back pattern and hold space

# Delete lines 2 and 3
2,3D

# Print pattern space before line 4 deletion
x  # Swap pattern and hold space
p  # Print the hold space (original pattern space)
x  # Swap back pattern and hold space

# Delete line 4
4D

# Print pattern space before appending hold space
x  # Swap pattern and hold space
p  # Print the hold space (original pattern space)
x  # Swap back pattern and hold space

# Append hold space to pattern space (debug the result)
G
# Print pattern space after appending hold space
p
