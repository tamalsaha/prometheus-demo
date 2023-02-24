package options

type Options struct {
	Address string `json:"address"`
	// +optional
	Username string `json:"username,omitempty"`
	// +optional
	PasswordPath string `json:"passwordPath,omitempty"`
	// +optional
	CredPath string `json:"credPath,omitempty"`
}

func (o *Options) Clone() *Options {
	return &Options{
		Address:      o.Address,
		Username:     o.Username,
		PasswordPath: o.PasswordPath,
		CredPath:     o.CredPath,
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Options) DeepCopyInto(out *Options) {
	*out = *in.Clone()
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Options.
func (in *Options) DeepCopy() *Options {
	if in == nil {
		return nil
	}
	out := new(Options)
	in.DeepCopyInto(out)
	return out
}
