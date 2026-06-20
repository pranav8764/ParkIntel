import onnx
from onnx import helper
import sys

def remove_zipmap(path):
    try:
        model = onnx.load(path)
    except Exception as e:
        print(f"Failed to load {path}: {e}")
        return False
        
    zipmap_node = None
    for node in model.graph.node:
        if node.op_type == "ZipMap":
            zipmap_node = node
            break
            
    if not zipmap_node:
        print(f"No ZipMap node found in {path}")
        return True
        
    # Get ZipMap input and output
    zip_in = zipmap_node.input[0]
    zip_out = zipmap_node.output[0]
    
    # Remove ZipMap node
    model.graph.node.remove(zipmap_node)
    
    # Change the output of the node before ZipMap to zip_out
    for node in model.graph.node:
        for idx, out_name in enumerate(node.output):
            if out_name == zip_in:
                node.output[idx] = zip_out
                
    # Update the type and shape of zip_out in graph outputs
    new_output = helper.make_tensor_value_info(
        zip_out,
        onnx.TensorProto.FLOAT,
        [None, 3] # [batch, num_classes]
    )
    
    for idx, out in enumerate(list(model.graph.output)):
        if out.name == zip_out:
            model.graph.output.remove(out)
            model.graph.output.insert(idx, new_output)
            break
            
    try:
        onnx.save(model, path)
        print(f"Successfully removed ZipMap from {path}")
        return True
    except Exception as e:
        print(f"Failed to save modified model {path}: {e}")
        return False

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python remove_zipmap.py <model_path>")
        sys.exit(1)
    success = remove_zipmap(sys.argv[1])
    sys.exit(0 if success else 1)
